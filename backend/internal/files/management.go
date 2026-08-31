package files

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type FileSummary struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	FolderID        *string   `json:"folderId,omitempty"`
	SizeBytes       int64     `json:"sizeBytes"`
	DeclaredMIME    string    `json:"declaredMime"`
	DetectedMIME    string    `json:"detectedMime"`
	WasDeduplicated bool      `json:"wasDeduplicated"`
	Visibility      string    `json:"visibility"`
	DownloadCount   int64     `json:"downloadCount"`
	UploadedAt      time.Time `json:"uploadedAt"`
	Tags            []string  `json:"tags"`
}

type FilePage struct {
	Items       []FileSummary `json:"items"`
	NextCursor  string        `json:"nextCursor,omitempty"`
	HasNextPage bool          `json:"hasNextPage"`
}

type FileDetail struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	UploaderName       string    `json:"uploaderName"`
	UploaderEmail      string    `json:"uploaderEmail"`
	SizeBytes          int64     `json:"sizeBytes"`
	DeclaredMIME       string    `json:"declaredMime"`
	DetectedMIME       string    `json:"detectedMime"`
	UploadedAt         time.Time `json:"uploadedAt"`
	WasDeduplicated    bool      `json:"wasDeduplicated"`
	BlobSizeBytes      int64     `json:"blobSizeBytes"`
	BlobReferenceCount int64     `json:"blobReferenceCount"`
	Visibility         string    `json:"visibility"`
	DownloadCount      int64     `json:"downloadCount"`
	PubliclyShared     bool      `json:"publiclyShared"`
}

type Folder struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parentId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

var ErrInvalidCursor = errors.New("invalid cursor")

type FileQuery struct {
	Cursor         string
	First          int
	Filename       string
	MIMETypes      []string
	MinSize        *int64
	MaxSize        *int64
	UploadedAfter  *time.Time
	UploadedBefore *time.Time
	UploaderName   string
	FolderID       string
	Tags           []string
}

func encodeCursor(t time.Time, id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(t.UTC().Format(time.RFC3339Nano) + "|" + id))
}

func decodeCursor(value string) (time.Time, string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}
	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", ErrInvalidCursor
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil || parts[1] == "" {
		return time.Time{}, "", ErrInvalidCursor
	}
	return t, parts[1], nil
}

func (s Service) List(ctx context.Context, ownerID string, q FileQuery) (FilePage, error) {
	first := q.First
	if first < 1 {
		first = 25
	}
	if first > 100 {
		first = 100
	}
	args := []any{ownerID}
	where := `f.owner_id = $1`
	add := func(value any) string { args = append(args, value); return fmt.Sprintf("$%d", len(args)) }
	if q.Filename != "" {
		// Keep the predicate aligned with files_name_trgm_idx so PostgreSQL can
		// use the trigram index for case-insensitive keyword search.
		where += ` AND lower(f.name) LIKE '%' || lower(` + add(q.Filename) + `) || '%'`
	}
	if len(q.MIMETypes) > 0 {
		where += ` AND f.detected_mime = ANY(` + add(q.MIMETypes) + `)`
	}
	if q.MinSize != nil {
		where += ` AND f.size_bytes >= ` + add(*q.MinSize)
	}
	if q.MaxSize != nil {
		where += ` AND f.size_bytes <= ` + add(*q.MaxSize)
	}
	if q.FolderID != "" {
		where += ` AND f.folder_id = ` + add(q.FolderID)
	}
	for _, tag := range q.Tags {
		where += ` AND EXISTS (SELECT 1 FROM file_tags ft JOIN tags t ON t.id=ft.tag_id WHERE ft.file_id=f.id AND lower(t.name)=lower(` + add(strings.TrimSpace(tag)) + `))`
	}
	if q.Cursor != "" {
		t, id, err := decodeCursor(q.Cursor)
		if err != nil {
			return FilePage{}, err
		}
		where += ` AND (f.created_at, f.id) < (` + add(t) + `, ` + add(id) + `)`
	}
	limitPlaceholder := add(first + 1)
	if q.UploadedAfter != nil {
		where += ` AND f.created_at >= ` + add(*q.UploadedAfter)
	}
	if q.UploadedBefore != nil {
		where += ` AND f.created_at <= ` + add(*q.UploadedBefore)
	}
	if q.UploaderName != "" {
		where += ` AND lower(u.display_name) LIKE '%' || lower(` + add(q.UploaderName) + `) || '%'`
	}
	query := fmt.Sprintf(`SELECT f.id::text, f.name, f.folder_id::text, f.size_bytes, f.declared_mime, f.detected_mime, f.was_deduplicated, f.visibility::text, f.download_count, f.created_at, COALESCE((SELECT json_agg(t.name ORDER BY t.name) FROM file_tags ft JOIN tags t ON t.id=ft.tag_id WHERE ft.file_id=f.id), '[]'::json) FROM files f JOIN users u ON u.id=f.uploaded_by WHERE %s ORDER BY f.created_at DESC, f.id DESC LIMIT %s`, where, limitPlaceholder)
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return FilePage{}, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()
	page := FilePage{Items: make([]FileSummary, 0)}
	for rows.Next() {
		var f FileSummary
		var tags []byte
		if err := rows.Scan(&f.ID, &f.Name, &f.FolderID, &f.SizeBytes, &f.DeclaredMIME, &f.DetectedMIME, &f.WasDeduplicated, &f.Visibility, &f.DownloadCount, &f.UploadedAt, &tags); err != nil {
			return FilePage{}, err
		}
		_ = json.Unmarshal(tags, &f.Tags)
		page.Items = append(page.Items, f)
	}
	if err := rows.Err(); err != nil {
		return FilePage{}, err
	}
	if len(page.Items) > first {
		page.HasNextPage = true
		last := page.Items[first-1]
		page.NextCursor = encodeCursor(last.UploadedAt, last.ID)
		page.Items = page.Items[:first]
	}
	return page, nil
}

func (s Service) Detail(ctx context.Context, ownerID, fileID string) (FileDetail, error) {
	var out FileDetail
	err := s.DB.QueryRowContext(ctx, `
		SELECT f.id::text, f.name, u.display_name, u.email::text, f.size_bytes,
		       f.declared_mime, f.detected_mime, f.created_at, f.was_deduplicated,
		       b.size_bytes, b.reference_count, f.visibility::text, f.download_count,
		       EXISTS (SELECT 1 FROM public_shares p
		               WHERE p.file_id = f.id AND p.revoked_at IS NULL
		                 AND (p.expires_at IS NULL OR p.expires_at > now()))
		FROM files f
		JOIN users u ON u.id = f.uploaded_by
		JOIN blobs b ON b.id = f.blob_id
		WHERE f.id = $1 AND f.owner_id = $2`, fileID, ownerID).
		Scan(&out.ID, &out.Name, &out.UploaderName, &out.UploaderEmail, &out.SizeBytes,
			&out.DeclaredMIME, &out.DetectedMIME, &out.UploadedAt, &out.WasDeduplicated,
			&out.BlobSizeBytes, &out.BlobReferenceCount, &out.Visibility,
			&out.DownloadCount, &out.PubliclyShared)
	if err == sql.ErrNoRows {
		return FileDetail{}, ErrFileNotFound
	}
	if err != nil {
		return FileDetail{}, fmt.Errorf("file detail: %w", err)
	}
	return out, nil
}

func (s Service) CreateFolder(ctx context.Context, ownerID, name, parentID string) (Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return Folder{}, errors.New("invalid folder name")
	}
	var f Folder
	err := s.DB.QueryRowContext(ctx, `INSERT INTO folders(owner_id,parent_id,name) VALUES($1,NULLIF($2,'')::uuid,$3) RETURNING id::text,name,parent_id::text,created_at`, ownerID, parentID, name).Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt)
	if err != nil {
		return Folder{}, fmt.Errorf("create folder: %w", err)
	}
	return f, nil
}

func (s Service) ListFolders(ctx context.Context, ownerID, parentID string) ([]Folder, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id::text,name,parent_id::text,created_at FROM folders WHERE owner_id=$1 AND parent_id IS NOT DISTINCT FROM NULLIF($2,'')::uuid ORDER BY lower(name),id`, ownerID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Folder, 0)
	for rows.Next() {
		var f Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s Service) Rename(ctx context.Context, ownerID, fileID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return errors.New("invalid filename")
	}
	r, err := s.DB.ExecContext(ctx, `UPDATE files SET name=$1,updated_at=now() WHERE id=$2 AND owner_id=$3`, name, fileID, ownerID)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrFileNotFound
	}
	return nil
}

func (s Service) Move(ctx context.Context, ownerID, fileID, folderID string) error {
	r, err := s.DB.ExecContext(ctx, `UPDATE files f SET folder_id=NULLIF($1,'')::uuid,updated_at=now() WHERE f.id=$2 AND f.owner_id=$3 AND (NULLIF($1,'') IS NULL OR EXISTS(SELECT 1 FROM folders d WHERE d.id=NULLIF($1,'')::uuid AND d.owner_id=$3))`, folderID, fileID, ownerID)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrFileNotFound
	}
	return nil
}

func (s Service) RenameFolder(ctx context.Context, ownerID, folderID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return errors.New("invalid folder name")
	}
	r, err := s.DB.ExecContext(ctx, `UPDATE folders SET name=$1,updated_at=now() WHERE id=$2 AND owner_id=$3`, name, folderID, ownerID)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrFolderNotFound
	}
	return nil
}

func (s Service) DeleteFolder(ctx context.Context, ownerID, folderID string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists, hasFiles, hasChildren bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM folders WHERE id=$1 AND owner_id=$2)`, folderID, ownerID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrFolderNotFound
	}
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM files WHERE folder_id=$1), EXISTS(SELECT 1 FROM folders WHERE parent_id=$1)`, folderID).Scan(&hasFiles, &hasChildren); err != nil {
		return err
	}
	if hasFiles || hasChildren {
		return ErrFolderNotEmpty
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM folders WHERE id=$1 AND owner_id=$2`, folderID, ownerID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s Service) SetTags(ctx context.Context, ownerID, fileID string, names []string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM files WHERE id=$1 AND owner_id=$2)`, fileID, ownerID).Scan(&exists); err != nil || !exists {
		return ErrFileNotFound
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM file_tags WHERE file_id=$1`, fileID); err != nil {
		return err
	}
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" || len(name) > 64 {
			return errors.New("invalid tag")
		}
		var tagID string
		if err = tx.QueryRowContext(ctx, `INSERT INTO tags(owner_id,name) VALUES($1,$2) ON CONFLICT(owner_id,name) DO UPDATE SET name=EXCLUDED.name RETURNING id::text`, ownerID, name).Scan(&tagID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO file_tags(file_id,tag_id) VALUES($1,$2)`, fileID, tagID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
