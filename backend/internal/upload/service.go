package upload

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/balkanid/file-vault/internal/storage"
)

var (
	ErrBlobNotReady  = errors.New("deduplicated blob is not ready")
	ErrQuotaExceeded = errors.New("logical storage quota exceeded")
	ErrEmptyFile     = errors.New("empty files are not allowed")
)

type QuotaError struct {
	Current   int64
	Requested int64
	Limit     int64
}

func (e QuotaError) Error() string { return ErrQuotaExceeded.Error() }
func (e QuotaError) Unwrap() error { return ErrQuotaExceeded }

type Service struct {
	DB    *sql.DB
	Store storage.Store
}

type Result struct {
	FileID       string
	Filename     string
	SizeBytes    int64
	DetectedMIME string
	Deduplicated bool
}

func (s Service) CreateFile(ctx context.Context, ownerID, filename, declaredMIME string, source io.Reader, maxBytes int64) (result Result, err error) {
	if s.DB == nil {
		return Result{}, errors.New("upload database is not configured")
	}
	name, err := NormalizeFilename(filename, 255)
	if err != nil {
		return Result{}, err
	}
	temp, err := s.Store.Stage(ctx, source, maxBytes)
	if err != nil {
		return Result{}, err
	}
	keepTemp := true
	defer func() {
		if keepTemp {
			_ = s.Store.Discard(temp)
		}
	}()
	detectedMIME, err := DetectFile(temp.Path)
	if err != nil {
		return Result{}, fmt.Errorf("detect MIME: %w", err)
	}
	if temp.Size == 0 {
		return Result{}, ErrEmptyFile
	}
	if err := ValidateMIME(name, declaredMIME, detectedMIME); err != nil {
		return Result{}, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("begin upload transaction: %w", err)
	}
	committed := false
	physicalCommitted := false
	var storageKey string
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
		if physicalCommitted && !committed {
			_ = s.Store.Delete(storageKey)
		}
	}()
	var quotaBytes, currentBytes int64
	if err := tx.QueryRowContext(ctx, `SELECT quota_bytes FROM users WHERE id = $1 FOR UPDATE`, ownerID).Scan(&quotaBytes); err != nil {
		return Result{}, fmt.Errorf("lock upload owner: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(size_bytes), 0) FROM files WHERE owner_id = $1`, ownerID).Scan(&currentBytes); err != nil {
		return Result{}, fmt.Errorf("calculate current quota usage: %w", err)
	}
	if currentBytes+temp.Size > quotaBytes {
		return Result{}, QuotaError{Current: currentBytes, Requested: temp.Size, Limit: quotaBytes}
	}

	digest := temp.SHA256[:]
	var blobID string
	var state string
	var isNew bool
	err = tx.QueryRowContext(ctx, `
		INSERT INTO blobs (sha256, size_bytes, detected_mime, storage_key, state)
		VALUES ($1, $2, $3, $4, 'pending')
		ON CONFLICT (sha256) DO NOTHING
		RETURNING id::text`, digest, temp.Size, detectedMIME, hex.EncodeToString(digest)).Scan(&blobID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `
			SELECT id::text, state::text, storage_key FROM blobs WHERE sha256 = $1 FOR UPDATE`, digest).
			Scan(&blobID, &state, &storageKey)
		if err != nil {
			return Result{}, fmt.Errorf("load deduplicated blob: %w", err)
		}
		if state != "ready" {
			return Result{}, ErrBlobNotReady
		}
	} else if err != nil {
		return Result{}, fmt.Errorf("insert blob: %w", err)
	} else {
		isNew = true
		storageKey = hex.EncodeToString(digest)
		if err := s.Store.Commit(temp, storageKey); err != nil {
			return Result{}, err
		}
		physicalCommitted = true
		if _, err := tx.ExecContext(ctx, `UPDATE blobs SET state = 'ready', updated_at = now() WHERE id = $1`, blobID); err != nil {
			return Result{}, fmt.Errorf("mark blob ready: %w", err)
		}
		keepTemp = false
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO files (owner_id, uploaded_by, blob_id, name, declared_mime, detected_mime, size_bytes, was_deduplicated)
		VALUES ($1, $1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text`, ownerID, blobID, name, declaredMIME, detectedMIME, temp.Size, !isNew).Scan(&result.FileID)
	if err != nil {
		return Result{}, fmt.Errorf("create file reference: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE blobs SET reference_count = reference_count + 1, updated_at = now() WHERE id = $1`, blobID); err != nil {
		return Result{}, fmt.Errorf("increment blob reference count: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit upload transaction: %w", err)
	}
	committed = true
	return Result{FileID: result.FileID, Filename: name, SizeBytes: temp.Size, DetectedMIME: detectedMIME, Deduplicated: !isNew}, nil
}

func (s Service) RemoveTemp(temp storage.TempObject) {
	if temp.Path != "" {
		_ = os.Remove(temp.Path)
	}
}
