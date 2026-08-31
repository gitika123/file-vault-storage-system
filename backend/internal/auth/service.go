package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/balkanid/file-vault/internal/auth/password"
	"github.com/balkanid/file-vault/internal/auth/policy"
	"github.com/balkanid/file-vault/internal/auth/session"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrSessionNotFound    = errors.New("session not found")
)

type User struct {
	ID          string
	Email       string
	DisplayName string
	Role        policy.Role
}

type Service struct {
	DB           *sql.DB
	SessionTTL   time.Duration
	CookieSecure bool
}

func (s Service) Authenticate(ctx context.Context, email, plainPassword string, now time.Time) (User, session.NewSession, error) {
	if s.DB == nil {
		return User{}, session.NewSession{}, errors.New("auth database is not configured")
	}
	var user User
	var role string
	var passwordHash string
	var disabledAt sql.NullTime
	err := s.DB.QueryRowContext(ctx, `
		SELECT id::text, email::text, display_name, role::text, password_hash, disabled_at
		FROM users WHERE email = $1`, strings.TrimSpace(strings.ToLower(email))).
		Scan(&user.ID, &user.Email, &user.DisplayName, &role, &passwordHash, &disabledAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, session.NewSession{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, session.NewSession{}, fmt.Errorf("load user: %w", err)
	}
	valid, err := password.Verify(passwordHash, plainPassword)
	if err != nil || !valid {
		return User{}, session.NewSession{}, ErrInvalidCredentials
	}
	if disabledAt.Valid {
		return User{}, session.NewSession{}, ErrAccountDisabled
	}
	user.Role = policy.Role(role)

	ttl := s.SessionTTL
	if ttl <= 0 {
		ttl = 8 * time.Hour
	}
	newSession, err := session.New(now.Add(ttl))
	if err != nil {
		return User{}, session.NewSession{}, err
	}
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token_hash, csrf_hash, expires_at)
		VALUES ($1, $2, $3, $4)`, user.ID, newSession.TokenHash[:], newSession.CSRFHash[:], newSession.ExpiresAt)
	if err != nil {
		return User{}, session.NewSession{}, fmt.Errorf("create session: %w", err)
	}
	return user, newSession, nil
}

func (s Service) Principal(ctx context.Context, rawToken string, now time.Time) (policy.Principal, [32]byte, error) {
	var principal policy.Principal
	var role string
	var csrfHash []byte
	if s.DB == nil || rawToken == "" {
		return principal, [32]byte{}, ErrSessionNotFound
	}
	tokenHash := sha256.Sum256([]byte(rawToken))
	var expiresAt time.Time
	err := s.DB.QueryRowContext(ctx, `
		SELECT s.user_id::text, u.role::text, s.csrf_hash, s.expires_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > $2 AND u.disabled_at IS NULL`, tokenHash[:], now).
		Scan(&principal.UserID, &role, &csrfHash, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return policy.Principal{}, [32]byte{}, ErrSessionNotFound
	}
	if err != nil {
		return policy.Principal{}, [32]byte{}, fmt.Errorf("load session: %w", err)
	}
	if len(csrfHash) != 32 {
		return policy.Principal{}, [32]byte{}, errors.New("invalid stored CSRF hash")
	}
	principal.Role = policy.Role(role)
	_, _ = s.DB.ExecContext(ctx, `UPDATE sessions SET last_seen_at = $1 WHERE token_hash = $2`, now, tokenHash[:])
	var hash [32]byte
	copy(hash[:], csrfHash)
	return principal, hash, nil
}

func (s Service) UserForToken(ctx context.Context, rawToken string, now time.Time) (User, [32]byte, error) {
	principal, csrfHash, err := s.Principal(ctx, rawToken, now)
	if err != nil {
		return User{}, [32]byte{}, err
	}
	var user User
	var role string
	err = s.DB.QueryRowContext(ctx, `
		SELECT id::text, email::text, display_name, role::text
		FROM users WHERE id = $1 AND disabled_at IS NULL`, principal.UserID).
		Scan(&user.ID, &user.Email, &user.DisplayName, &role)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, [32]byte{}, ErrSessionNotFound
	}
	if err != nil {
		return User{}, [32]byte{}, fmt.Errorf("load current user: %w", err)
	}
	user.Role = policy.Role(role)
	return user, csrfHash, nil
}

func (s Service) Revoke(ctx context.Context, rawToken string) error {
	if s.DB == nil || rawToken == "" {
		return nil
	}
	tokenHash := sha256.Sum256([]byte(rawToken))
	_, err := s.DB.ExecContext(ctx, `UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash[:])
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}
