package files

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/balkanid/file-vault/internal/auth/policy"
	"github.com/balkanid/file-vault/internal/storage"
)

var (
	ErrFileNotFound   = errors.New("file not found")
	ErrNotUploader    = errors.New("only the uploader can delete this file")
	ErrFolderNotFound = errors.New("folder not found")
	ErrFolderNotEmpty = errors.New("folder must be empty before deletion")
)

type Service struct {
	DB     *sql.DB
	Store  storage.Store
	Events *DownloadEventHub
}

type StorageStats struct {
	OriginalBytes     int64   `json:"originalBytes"`
	DeduplicatedBytes int64   `json:"deduplicatedBytes"`
	SavingsBytes      int64   `json:"savingsBytes"`
	SavingsPercent    float64 `json:"savingsPercent"`
	QuotaBytes        int64   `json:"quotaBytes"`
}

type DeleteResult struct {
	CleanupPending bool `json:"cleanupPending"`
}

func (s Service) StorageStats(ctx context.Context, ownerID string) (StorageStats, error) {
	var stats StorageStats
	err := s.DB.QueryRowContext(ctx, `
		WITH user_files AS (
			SELECT blob_id, size_bytes FROM files WHERE owner_id = $1
		), distinct_blobs AS (
			SELECT DISTINCT blob_id FROM user_files
		)
		SELECT
			COALESCE((SELECT SUM(size_bytes) FROM user_files), 0),
			COALESCE((SELECT SUM(b.size_bytes) FROM blobs b JOIN distinct_blobs d ON d.blob_id = b.id), 0),
			u.quota_bytes
		FROM users u WHERE u.id = $1`, ownerID).
		Scan(&stats.OriginalBytes, &stats.DeduplicatedBytes, &stats.QuotaBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return StorageStats{}, ErrFileNotFound
	}
	if err != nil {
		return StorageStats{}, fmt.Errorf("calculate storage statistics: %w", err)
	}
	stats.SavingsBytes = stats.OriginalBytes - stats.DeduplicatedBytes
	if stats.OriginalBytes > 0 {
		stats.SavingsPercent = float64(stats.SavingsBytes) * 100 / float64(stats.OriginalBytes)
	}
	return stats, nil
}

func (s Service) Delete(ctx context.Context, principal *policy.Principal, fileID string) (DeleteResult, error) {
	if err := policy.RequireAuthenticated(principal); err != nil {
		return DeleteResult{}, err
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return DeleteResult{}, fmt.Errorf("begin deletion transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	var uploaderID, blobID, storageKey string
	var references int64
	err = tx.QueryRowContext(ctx, `
		SELECT f.uploaded_by::text, f.blob_id::text, b.storage_key, b.reference_count
		FROM files f JOIN blobs b ON b.id = f.blob_id
		WHERE f.id = $1
		FOR UPDATE OF f, b`, fileID).
		Scan(&uploaderID, &blobID, &storageKey, &references)
	if errors.Is(err, sql.ErrNoRows) {
		return DeleteResult{}, ErrFileNotFound
	}
	if err != nil {
		return DeleteResult{}, fmt.Errorf("load file for deletion: %w", err)
	}
	if uploaderID != principal.UserID {
		return DeleteResult{}, ErrNotUploader
	}
	if references <= 0 {
		return DeleteResult{}, errors.New("blob reference count is already invalid")
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE id = $1`, fileID); err != nil {
		return DeleteResult{}, fmt.Errorf("delete file reference: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE blobs SET reference_count = reference_count - 1,
		state = CASE WHEN reference_count = 1 THEN 'pending_delete'::blob_state ELSE state END,
		updated_at = now() WHERE id = $1`, blobID); err != nil {
		return DeleteResult{}, fmt.Errorf("decrement blob reference count: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return DeleteResult{}, fmt.Errorf("commit deletion transaction: %w", err)
	}
	committed = true
	if references > 1 {
		return DeleteResult{}, nil
	}
	cleanupPending := false
	if err := s.Store.Delete(storageKey); err != nil && !errors.Is(err, os.ErrNotExist) {
		cleanupPending = true
	} else {
		if _, err := s.DB.ExecContext(ctx, `DELETE FROM blobs WHERE id = $1 AND reference_count = 0`, blobID); err != nil {
			cleanupPending = true
		}
	}
	return DeleteResult{CleanupPending: cleanupPending}, nil
}
