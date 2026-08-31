package files

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/balkanid/file-vault/internal/auth/policy"
)

var ErrContentForbidden = errors.New("content access forbidden")

type contentRecord struct {
	FileID, Name, MIME, StorageKey string
	Size                           int64
	ShareID                        *string
}

func (s Service) Content(ctx context.Context, principal *policy.Principal, fileID, token string, preview bool) (contentRecord, error) {
	var rec contentRecord
	var owner, permission string
	var expires *time.Time
	if token != "" {
		sum := sha256.Sum256([]byte(token))
		err := s.DB.QueryRowContext(ctx, `SELECT f.id::text,f.name,f.detected_mime,b.storage_key,b.size_bytes,p.id::text,p.expires_at FROM public_shares p JOIN files f ON f.id=p.file_id JOIN blobs b ON b.id=f.blob_id WHERE p.token_hash=$1 AND p.revoked_at IS NULL AND (p.expires_at IS NULL OR p.expires_at>now())`, sum[:]).Scan(&rec.FileID, &rec.Name, &rec.MIME, &rec.StorageKey, &rec.Size, &rec.ShareID, &expires)
		if errors.Is(err, sql.ErrNoRows) {
			return contentRecord{}, ErrContentForbidden
		}
		if err != nil {
			return contentRecord{}, err
		}
		_ = expires
	} else {
		if principal == nil || principal.UserID == "" {
			return contentRecord{}, policy.ErrUnauthenticated
		}
		err := s.DB.QueryRowContext(ctx, `SELECT f.id::text,f.name,f.detected_mime,b.storage_key,b.size_bytes,f.owner_id::text FROM files f JOIN blobs b ON b.id=f.blob_id WHERE f.id=$1`, fileID).Scan(&rec.FileID, &rec.Name, &rec.MIME, &rec.StorageKey, &rec.Size, &owner)
		if errors.Is(err, sql.ErrNoRows) {
			return contentRecord{}, ErrFileNotFound
		}
		if err != nil {
			return contentRecord{}, err
		}
		if owner != principal.UserID && principal.Role != policy.RoleAdmin {
			var folderID sql.NullString
			_ = s.DB.QueryRowContext(ctx, `SELECT folder_id::text FROM files WHERE id=$1`, fileID).Scan(&folderID)
			err = s.DB.QueryRowContext(ctx, `SELECT permission::text FROM user_shares WHERE recipient_id=$2 AND (file_id=$1 OR folder_id=NULLIF($3,'')::uuid) ORDER BY file_id NULLS LAST LIMIT 1`, fileID, principal.UserID, folderID.String).Scan(&permission)
			if err != nil || (permission != "download" && !(preview && permission == "view")) {
				return contentRecord{}, ErrContentForbidden
			}
		}
	}
	if preview && !previewableMIME(rec.MIME) {
		return contentRecord{}, ErrContentForbidden
	}
	if !preview {
		var ownerID string
		var downloadCount int64
		if err := s.DB.QueryRowContext(ctx, `UPDATE files SET download_count=download_count+1,updated_at=now() WHERE id=$1 RETURNING owner_id::text, download_count`, rec.FileID).Scan(&ownerID, &downloadCount); err != nil {
			return contentRecord{}, fmt.Errorf("increment download count: %w", err)
		}
		s.Events.Publish(DownloadEvent{OwnerID: ownerID, FileID: rec.FileID, DownloadCount: downloadCount})
		if _, err := s.DB.ExecContext(ctx, `INSERT INTO download_events(file_id,blob_id,actor_id,public_share_id) SELECT f.id,f.blob_id,$2::uuid,$3::uuid FROM files f WHERE f.id=$1`, rec.FileID, principalID(principal), rec.ShareID); err != nil {
			return contentRecord{}, err
		}
		if _, err := s.DB.ExecContext(ctx, `INSERT INTO audit_events(actor_id,action,resource_type,resource_id,metadata) VALUES($1,'download','file',$2,jsonb_build_object('publicShare', $3::boolean))`, principalID(principal), rec.FileID, rec.ShareID != nil); err != nil {
			return contentRecord{}, err
		}
	}
	return rec, nil
}

func previewableMIME(mimeType string) bool {
	return mimeType == "application/pdf" || strings.HasPrefix(mimeType, "image/") || strings.HasPrefix(mimeType, "audio/") || strings.HasPrefix(mimeType, "video/") || strings.HasPrefix(mimeType, "text/")
}

func principalID(p *policy.Principal) any {
	if p == nil {
		return nil
	}
	return p.UserID
}

func (s Service) ServeContent(ctx context.Context, w http.ResponseWriter, r *http.Request, principal *policy.Principal, fileID, token string, preview bool) error {
	rec, err := s.Content(ctx, principal, fileID, token, preview)
	if err != nil {
		return err
	}
	f, err := s.Store.Open(rec.StorageKey)
	if err != nil {
		return err
	}
	defer f.Close()
	w.Header().Set("Content-Type", rec.MIME)
	disp := "attachment"
	if preview {
		disp = "inline"
	}
	w.Header().Set("Content-Disposition", disp+`; filename="`+strings.ReplaceAll(rec.Name, "\"", "_")+`"`)
	http.ServeContent(w, r, rec.Name, time.Time{}, f)
	return nil
}

func (s Service) ServePublicFolderAsset(ctx context.Context, w http.ResponseWriter, r *http.Request, token, fileID string, preview bool) error {
	sum := sha256.Sum256([]byte(token))
	var rec contentRecord
	var expires *time.Time
	err := s.DB.QueryRowContext(ctx, `SELECT f.id::text,f.name,f.detected_mime,b.storage_key,b.size_bytes,p.id::text,p.expires_at FROM public_shares p JOIN files f ON f.folder_id=p.folder_id JOIN blobs b ON b.id=f.blob_id WHERE p.token_hash=$1 AND f.id=$2 AND p.revoked_at IS NULL AND (p.expires_at IS NULL OR p.expires_at>now())`, sum[:], fileID).Scan(&rec.FileID, &rec.Name, &rec.MIME, &rec.StorageKey, &rec.Size, &rec.ShareID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrContentForbidden
	}
	if err != nil {
		return err
	}
	if preview && rec.MIME != "application/pdf" && !strings.HasPrefix(rec.MIME, "image/") {
		return ErrContentForbidden
	}
	return s.serveStoredContent(ctx, w, r, rec, preview)
}

func (s Service) serveStoredContent(ctx context.Context, w http.ResponseWriter, r *http.Request, rec contentRecord, preview bool) error {
	f, err := s.Store.Open(rec.StorageKey)
	if err != nil {
		return err
	}
	defer f.Close()
	if !preview {
		var ownerID string
		var downloadCount int64
		if err := s.DB.QueryRowContext(ctx, `UPDATE files SET download_count=download_count+1,updated_at=now() WHERE id=$1 RETURNING owner_id::text, download_count`, rec.FileID).Scan(&ownerID, &downloadCount); err != nil {
			return fmt.Errorf("increment download count: %w", err)
		}
		s.Events.Publish(DownloadEvent{OwnerID: ownerID, FileID: rec.FileID, DownloadCount: downloadCount})
		if _, err := s.DB.ExecContext(ctx, `INSERT INTO download_events(file_id,blob_id,actor_id,public_share_id) SELECT f.id,f.blob_id,NULL::uuid,$2::uuid FROM files f WHERE f.id=$1`, rec.FileID, rec.ShareID); err != nil {
			return err
		}
		if _, err := s.DB.ExecContext(ctx, `INSERT INTO audit_events(actor_id,action,resource_type,resource_id,metadata) VALUES(NULL::uuid,'download','file',$1,jsonb_build_object('publicShare', true))`, rec.FileID); err != nil {
			return err
		}
	}
	w.Header().Set("Content-Type", rec.MIME)
	disp := "attachment"
	if preview {
		disp = "inline"
	}
	w.Header().Set("Content-Disposition", disp+`; filename="`+strings.ReplaceAll(rec.Name, `"`, "_")+`"`)
	http.ServeContent(w, r, rec.Name, time.Time{}, f)
	return nil
}

func (s Service) PublicResource(ctx context.Context, token string) (string, error) {
	sum := sha256.Sum256([]byte(token))
	var fileID, fileName, mimeType string
	var folderID sql.NullString
	err := s.DB.QueryRowContext(ctx, `SELECT COALESCE(f.id::text,''),COALESCE(f.name,''),COALESCE(f.detected_mime,''),p.folder_id::text FROM public_shares p LEFT JOIN files f ON f.id=p.file_id WHERE p.token_hash=$1 AND p.revoked_at IS NULL AND (p.expires_at IS NULL OR p.expires_at>now())`, sum[:]).Scan(&fileID, &fileName, &mimeType, &folderID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrContentForbidden
	}
	if err != nil {
		return "", err
	}
	if fileID != "" {
		return fmt.Sprintf(`<!doctype html><meta charset="utf-8"><title>%s - BalkanID</title><style>body{font:16px system-ui;margin:40px;max-width:900px;color:#17334d}iframe{width:100%%;height:70vh;border:1px solid #dfe7ee;border-radius:8px}.actions{display:flex;gap:12px}a{color:#1a7c82}</style><h1>%s</h1><p>%s</p><div class="actions"><a href="/public/%s/download?preview=1">Open preview</a><a href="/public/%s/download">Download</a></div><iframe src="/public/%s/download?preview=1" title="File preview"></iframe>`, html.EscapeString(fileName), html.EscapeString(fileName), html.EscapeString(mimeType), html.EscapeString(token), html.EscapeString(token), html.EscapeString(token)), nil
	}
	if !folderID.Valid {
		return "", ErrContentForbidden
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT id::text,name,detected_mime FROM files WHERE folder_id=$1 ORDER BY lower(name),id`, folderID.String)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	body := fmt.Sprintf(`<!doctype html><meta charset="utf-8"><title>Shared folder - BalkanID</title><style>body{font:16px system-ui;margin:40px;max-width:900px;color:#17334d}li{margin:14px 0}a{color:#1a7c82}</style><h1>Shared folder</h1><p>Files shared from this folder:</p><ul>`)
	for rows.Next() {
		var id, name, mime string
		if err := rows.Scan(&id, &name, &mime); err != nil {
			return "", err
		}
		body += fmt.Sprintf(`<li><strong>%s</strong> <small>%s</small> — <a href="/public/%s/download?fileId=%s&preview=1">Preview</a> · <a href="/public/%s/download?fileId=%s">Download</a></li>`, html.EscapeString(name), html.EscapeString(mime), html.EscapeString(token), html.EscapeString(id), html.EscapeString(token), html.EscapeString(id))
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return body + `</ul>`, nil
}
