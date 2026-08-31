package files

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrShareNotFound     = errors.New("share not found")
	ErrShareForbidden    = errors.New("share owner permission required")
	ErrRecipientNotFound = errors.New("recipient not found")
)

type PublicShare struct {
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type ShareAccess struct {
	ID             string `json:"id"`
	RecipientEmail string `json:"recipientEmail"`
	RecipientName  string `json:"recipientName"`
	Permission     string `json:"permission"`
}

func newShareToken() (string, []byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, sum[:], nil
}

func (s Service) CreatePublicShare(ctx context.Context, ownerID, fileID string, expiresAt *time.Time) (PublicShare, error) {
	token, hash, err := newShareToken()
	if err != nil {
		return PublicShare{}, err
	}
	var exists bool
	if err = s.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM files WHERE id=$1 AND owner_id=$2)`, fileID, ownerID).Scan(&exists); err != nil {
		return PublicShare{}, err
	}
	if !exists {
		return PublicShare{}, ErrFileNotFound
	}
	var out PublicShare
	err = s.DB.QueryRowContext(ctx, `INSERT INTO public_shares(file_id,token_hash,created_by,expires_at) VALUES($1,$2,$3,$4) RETURNING expires_at`, fileID, hash, ownerID, expiresAt).Scan(&out.ExpiresAt)
	if err != nil {
		return PublicShare{}, fmt.Errorf("create public share: %w", err)
	}
	out.Token = token
	return out, nil
}

func (s Service) CreatePublicFolderShare(ctx context.Context, ownerID, folderID string, expiresAt *time.Time) (PublicShare, error) {
	token, hash, err := newShareToken()
	if err != nil {
		return PublicShare{}, err
	}
	var exists bool
	if err = s.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM folders WHERE id=$1 AND owner_id=$2)`, folderID, ownerID).Scan(&exists); err != nil {
		return PublicShare{}, err
	}
	if !exists {
		return PublicShare{}, ErrFolderNotFound
	}
	var out PublicShare
	err = s.DB.QueryRowContext(ctx, `INSERT INTO public_shares(folder_id,token_hash,created_by,expires_at) VALUES($1,$2,$3,$4) RETURNING expires_at`, folderID, hash, ownerID, expiresAt).Scan(&out.ExpiresAt)
	if err != nil {
		return PublicShare{}, fmt.Errorf("create public folder share: %w", err)
	}
	out.Token = token
	return out, nil
}

func (s Service) RevokePublicShare(ctx context.Context, ownerID, fileID string) error {
	r, err := s.DB.ExecContext(ctx, `UPDATE public_shares p SET revoked_at=now() FROM files f WHERE p.file_id=$1 AND f.id=p.file_id AND f.owner_id=$2 AND p.revoked_at IS NULL`, fileID, ownerID)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrShareNotFound
	}
	return nil
}

func (s Service) RevokePublicFolderShare(ctx context.Context, ownerID, folderID string) error {
	r, err := s.DB.ExecContext(ctx, `UPDATE public_shares p SET revoked_at=now() FROM folders f WHERE p.folder_id=$1 AND f.id=p.folder_id AND f.owner_id=$2 AND p.revoked_at IS NULL`, folderID, ownerID)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrShareNotFound
	}
	return nil
}

func (s Service) ListShareAccess(ctx context.Context, ownerID, resourceID string, folder bool) ([]ShareAccess, error) {
	column := "file_id"
	ownerTable := "files"
	if folder {
		column, ownerTable = "folder_id", "folders"
	}
	query := fmt.Sprintf(`SELECT us.id::text, u.email, u.display_name, us.permission::text FROM user_shares us JOIN users u ON u.id=us.recipient_id JOIN %s r ON r.id=us.%s WHERE r.owner_id=$1 AND us.%s=$2 ORDER BY lower(u.display_name), lower(u.email)`, ownerTable, column, column)
	rows, err := s.DB.QueryContext(ctx, query, ownerID, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	access := make([]ShareAccess, 0)
	for rows.Next() {
		var item ShareAccess
		if err := rows.Scan(&item.ID, &item.RecipientEmail, &item.RecipientName, &item.Permission); err != nil {
			return nil, err
		}
		access = append(access, item)
	}
	return access, rows.Err()
}

func (s Service) RevokeDirectShare(ctx context.Context, ownerID, shareID string) error {
	r, err := s.DB.ExecContext(ctx, `DELETE FROM user_shares us USING files f WHERE us.id=$1 AND us.file_id=f.id AND f.owner_id=$2`, shareID, ownerID)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrShareNotFound
	}
	return nil
}

func (s Service) RevokeDirectFolderShare(ctx context.Context, ownerID, shareID string) error {
	r, err := s.DB.ExecContext(ctx, `DELETE FROM user_shares us USING folders f WHERE us.id=$1 AND us.folder_id=f.id AND f.owner_id=$2`, shareID, ownerID)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrShareNotFound
	}
	return nil
}

func (s Service) CreateDirectShare(ctx context.Context, ownerID, fileID, recipientEmail, permission string) error {
	permission = strings.ToLower(strings.TrimSpace(permission))
	if permission != "view" && permission != "download" {
		return errors.New("invalid permission")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var recipient string
	if err = tx.QueryRowContext(ctx, `SELECT id::text FROM users WHERE email=$1 AND disabled_at IS NULL`, strings.TrimSpace(recipientEmail)).Scan(&recipient); errors.Is(err, sql.ErrNoRows) {
		return ErrRecipientNotFound
	} else if err != nil {
		return err
	}
	if recipient == ownerID {
		return errors.New("cannot share with self")
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM files WHERE id=$1 AND owner_id=$2)`, fileID, ownerID).Scan(&exists); err != nil || !exists {
		return ErrFileNotFound
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_shares(file_id,recipient_id,granted_by,permission) VALUES($1,$2,$3,$4::share_permission) ON CONFLICT DO NOTHING`, fileID, recipient, ownerID, permission); err != nil {
		return err
	}
	return tx.Commit()
}

func (s Service) CreateDirectFolderShare(ctx context.Context, ownerID, folderID, recipientEmail, permission string) error {
	permission = strings.ToLower(strings.TrimSpace(permission))
	if permission != "view" && permission != "download" {
		return errors.New("invalid permission")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var recipient string
	if err = tx.QueryRowContext(ctx, `SELECT id::text FROM users WHERE email=$1 AND disabled_at IS NULL`, strings.TrimSpace(recipientEmail)).Scan(&recipient); errors.Is(err, sql.ErrNoRows) {
		return ErrRecipientNotFound
	} else if err != nil {
		return err
	}
	if recipient == ownerID {
		return errors.New("cannot share with self")
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM folders WHERE id=$1 AND owner_id=$2)`, folderID, ownerID).Scan(&exists); err != nil || !exists {
		return ErrFolderNotFound
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_shares(folder_id,recipient_id,granted_by,permission) VALUES($1,$2,$3,$4::share_permission) ON CONFLICT DO NOTHING`, folderID, recipient, ownerID, permission); err != nil {
		return err
	}
	return tx.Commit()
}
