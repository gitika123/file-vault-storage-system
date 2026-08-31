package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	CookieName = "balkanid_session"
	CSRFCookie = "balkanid_csrf"
	CSRFHeader = "X-CSRF-Token"
)

var ErrTokenGeneration = errors.New("unable to generate secure token")

type NewSession struct {
	RawToken  string
	TokenHash [32]byte
	CSRFToken string
	CSRFHash  [32]byte
	ExpiresAt time.Time
}

func New(expiresAt time.Time) (NewSession, error) {
	raw, err := secureToken(32)
	if err != nil {
		return NewSession{}, err
	}
	csrf, err := secureToken(32)
	if err != nil {
		return NewSession{}, err
	}
	return NewSession{RawToken: raw, TokenHash: sha256.Sum256([]byte(raw)), CSRFToken: csrf, CSRFHash: sha256.Sum256([]byte(csrf)), ExpiresAt: expiresAt}, nil
}

func ValidateCSRF(r *http.Request, expectedHash [32]byte) error {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return nil
	}
	token := r.Header.Get(CSRFHeader)
	if token == "" {
		return fmt.Errorf("%s header is required", CSRFHeader)
	}
	actual := sha256.Sum256([]byte(token))
	if actual != expectedHash {
		return errors.New("invalid CSRF token")
	}
	return nil
}

func SetCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: token, Path: "/", Expires: expiresAt, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
}

func ClearCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
	http.SetCookie(w, &http.Cookie{Name: CSRFCookie, Value: "", Path: "/", MaxAge: -1, Secure: secure, SameSite: http.SameSiteLaxMode})
}

func SetCSRFCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: CSRFCookie, Value: token, Path: "/", Expires: expiresAt, Secure: secure, SameSite: http.SameSiteLaxMode})
}

func secureToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
