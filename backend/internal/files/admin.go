package files

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type AdminStats struct {
	Users         int64 `json:"users"`
	Files         int64 `json:"files"`
	LogicalBytes  int64 `json:"logicalBytes"`
	PhysicalBytes int64 `json:"physicalBytes"`
	Downloads     int64 `json:"downloads"`
}

type AdminFile struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	UploaderName    string    `json:"uploaderName"`
	UploaderEmail   string    `json:"uploaderEmail"`
	SizeBytes       int64     `json:"sizeBytes"`
	DetectedMIME    string    `json:"detectedMime"`
	UploadedAt      time.Time `json:"uploadedAt"`
	WasDeduplicated bool      `json:"wasDeduplicated"`
	Visibility      string    `json:"visibility"`
	DownloadCount   int64     `json:"downloadCount"`
}

type AdminFilePage struct {
	Items       []AdminFile `json:"items"`
	NextCursor  string      `json:"nextCursor,omitempty"`
	HasNextPage bool        `json:"hasNextPage"`
}

func (s Service) AdminFiles(ctx context.Context, first int, cursor string) (AdminFilePage, error) {
	if first < 1 || first > 100 {
		first = 50
	}
	args := []any{first + 1}
	where := ""
	if cursor != "" {
		t, id, err := decodeCursor(cursor)
		if err != nil {
			return AdminFilePage{}, ErrInvalidCursor
		}
		args = append(args, t, id)
		where = "WHERE (f.created_at, f.id) < ($2, $3)"
	}
	rows, err := s.DB.QueryContext(ctx, `
		SELECT f.id::text, f.name, u.display_name, u.email::text, f.size_bytes,
		       f.detected_mime, f.created_at, f.was_deduplicated,
		       f.visibility::text, f.download_count
		FROM files f JOIN users u ON u.id=f.uploaded_by `+where+`
		ORDER BY f.created_at DESC, f.id DESC LIMIT $1`, args...)
	if err != nil {
		return AdminFilePage{}, fmt.Errorf("admin files: %w", err)
	}
	defer rows.Close()
	page := AdminFilePage{Items: make([]AdminFile, 0)}
	for rows.Next() {
		var file AdminFile
		if err := rows.Scan(&file.ID, &file.Name, &file.UploaderName, &file.UploaderEmail, &file.SizeBytes, &file.DetectedMIME, &file.UploadedAt, &file.WasDeduplicated, &file.Visibility, &file.DownloadCount); err != nil {
			return AdminFilePage{}, err
		}
		page.Items = append(page.Items, file)
	}
	if err := rows.Err(); err != nil {
		return AdminFilePage{}, err
	}
	if len(page.Items) > first {
		page.HasNextPage = true
		last := page.Items[first-1]
		page.NextCursor = encodeCursor(last.UploadedAt, last.ID)
		page.Items = page.Items[:first]
	}
	return page, nil
}

func (s Service) AdminStats(ctx context.Context) (AdminStats, error) {
	var out AdminStats
	err := s.DB.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM users WHERE disabled_at IS NULL),(SELECT count(*) FROM files),(SELECT coalesce(sum(size_bytes),0) FROM files),(SELECT coalesce(sum(size_bytes),0) FROM blobs WHERE state='ready'),(SELECT count(*) FROM download_events)`).Scan(&out.Users, &out.Files, &out.LogicalBytes, &out.PhysicalBytes, &out.Downloads)
	if err == sql.ErrNoRows {
		return out, nil
	}
	if err != nil {
		return out, fmt.Errorf("admin stats: %w", err)
	}
	return out, nil
}
