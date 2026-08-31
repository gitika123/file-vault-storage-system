package files

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/balkanid/file-vault/internal/auth"
	"github.com/balkanid/file-vault/internal/auth/policy"
)

type HTTPHandler struct {
	Service Service
}

func (h HTTPHandler) DownloadEvents(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || h.Service.Events == nil {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, 500, "INTERNAL", "Streaming is unavailable.")
		return
	}
	ch, cancel := h.Service.Events.Subscribe()
	defer cancel()
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case event, open := <-ch:
			if !open {
				return
			}
			if p.Role != policy.RoleAdmin && event.OwnerID != p.UserID {
				continue
			}
			payload, _ := json.Marshal(event)
			_, _ = w.Write([]byte("event: download\ndata: " + string(payload) + "\n\n"))
			flusher.Flush()
		}
	}
}

func (h HTTPHandler) Stats(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	stats, err := h.Service.StorageStats(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Unable to calculate storage statistics.")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h HTTPHandler) AdminStats(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.Role != policy.RoleAdmin {
		writeError(w, 403, "FORBIDDEN", "Administrator access required.")
		return
	}
	out, err := h.Service.AdminStats(r.Context())
	if err != nil {
		writeError(w, 500, "INTERNAL", "Unable to calculate admin statistics.")
		return
	}
	writeJSON(w, 200, out)
}

func (h HTTPHandler) AdminFiles(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.Role != policy.RoleAdmin {
		writeError(w, 403, "FORBIDDEN", "Administrator access required.")
		return
	}
	first, _ := strconv.Atoi(r.URL.Query().Get("first"))
	files, err := h.Service.AdminFiles(r.Context(), first, r.URL.Query().Get("after"))
	if errors.Is(err, ErrInvalidCursor) {
		writeError(w, 400, "INVALID_INPUT", "Cursor is invalid.")
		return
	}
	if err != nil {
		writeError(w, 500, "INTERNAL", "Unable to list administrator files.")
		return
	}
	writeJSON(w, 200, files)
}

func (h HTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	fileID := r.PathValue("id")
	result, err := h.Service.Delete(r.Context(), principal, fileID)
	if err != nil {
		status, code, message := http.StatusInternalServerError, "INTERNAL", "Unable to delete file."
		switch {
		case errors.Is(err, ErrFileNotFound):
			status, code, message = http.StatusNotFound, "NOT_FOUND", "File not found."
		case errors.Is(err, ErrNotUploader), errors.Is(err, policy.ErrForbidden):
			status, code, message = http.StatusForbidden, "FORBIDDEN", "Only the uploader can delete this file."
		}
		writeError(w, status, code, message)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "cleanupPending": result.CleanupPending})
}

func (h HTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	values := r.URL.Query()
	first, _ := strconv.Atoi(values.Get("first"))
	q := FileQuery{First: first, Cursor: values.Get("after"), Filename: values.Get("filename"), FolderID: values.Get("folderId")}
	var parseErr bool
	if raw := values.Get("uploadedAfter"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			q.UploadedAfter = &t
		} else {
			parseErr = true
		}
	}
	if raw := values.Get("uploadedBefore"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			q.UploadedBefore = &t
		} else {
			parseErr = true
		}
	}
	q.UploaderName = values.Get("uploaderName")
	if q.UploadedAfter != nil && q.UploadedBefore != nil && q.UploadedAfter.After(*q.UploadedBefore) {
		parseErr = true
	}
	if parseErr {
		writeError(w, 400, "INVALID_INPUT", "Date range is invalid.")
		return
	}
	for _, raw := range strings.Split(values.Get("mime"), ",") {
		if raw = strings.TrimSpace(raw); raw != "" {
			q.MIMETypes = append(q.MIMETypes, raw)
		}
	}
	for _, raw := range strings.Split(values.Get("tag"), ",") {
		if raw = strings.TrimSpace(raw); raw != "" {
			q.Tags = append(q.Tags, raw)
		}
	}
	if raw := values.Get("minSizeBytes"); raw != "" {
		if n, e := strconv.ParseInt(raw, 10, 64); e == nil {
			q.MinSize = &n
		}
	}
	if raw := values.Get("maxSizeBytes"); raw != "" {
		if n, e := strconv.ParseInt(raw, 10, 64); e == nil {
			q.MaxSize = &n
		}
	}
	page, err := h.Service.List(r.Context(), p.UserID, q)
	if err != nil {
		if errors.Is(err, ErrInvalidCursor) {
			writeError(w, 400, "INVALID_INPUT", "Cursor is invalid.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to list files.")
		}
		return
	}
	writeJSON(w, 200, page)
}

func (h HTTPHandler) Detail(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	out, err := h.Service.Detail(r.Context(), p.UserID, r.PathValue("id"))
	if errors.Is(err, ErrFileNotFound) {
		writeError(w, 404, "NOT_FOUND", "File not found.")
		return
	}
	if err != nil {
		writeError(w, 500, "INTERNAL", "Unable to load file details.")
		return
	}
	writeJSON(w, 200, out)
}

func (h HTTPHandler) Folders(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	out, err := h.Service.ListFolders(r.Context(), p.UserID, r.URL.Query().Get("parentId"))
	if err != nil {
		writeError(w, 500, "INTERNAL", "Unable to list folders.")
		return
	}
	writeJSON(w, 200, out)
}

func (h HTTPHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var in struct {
		Name     string `json:"name"`
		ParentID string `json:"parentId"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in) != nil {
		writeError(w, 400, "INVALID_INPUT", "Invalid folder payload.")
		return
	}
	f, err := h.Service.CreateFolder(r.Context(), p.UserID, in.Name, in.ParentID)
	if err != nil {
		writeError(w, 400, "INVALID_INPUT", "Folder name or parent is invalid.")
		return
	}
	writeJSON(w, 201, f)
}

func (h HTTPHandler) RenameFolder(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in) != nil {
		writeError(w, 400, "INVALID_INPUT", "Invalid folder payload.")
		return
	}
	if err := h.Service.RenameFolder(r.Context(), p.UserID, r.PathValue("id"), in.Name); err != nil {
		if errors.Is(err, ErrFolderNotFound) {
			writeError(w, 404, "NOT_FOUND", "Folder not found.")
		} else {
			writeError(w, 400, "INVALID_INPUT", "Folder name is invalid.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"updated": true})
}

func (h HTTPHandler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	if err := h.Service.DeleteFolder(r.Context(), p.UserID, r.PathValue("id")); err != nil {
		switch {
		case errors.Is(err, ErrFolderNotFound):
			writeError(w, 404, "NOT_FOUND", "Folder not found.")
		case errors.Is(err, ErrFolderNotEmpty):
			writeError(w, 409, "FOLDER_NOT_EMPTY", "Move or remove the folder contents before deleting this folder.")
		default:
			writeError(w, 500, "INTERNAL", "Unable to delete folder.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"deleted": true})
}

func (h HTTPHandler) Rename(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in) != nil {
		writeError(w, 400, "INVALID_INPUT", "Invalid rename payload.")
		return
	}
	if err := h.Service.Rename(r.Context(), p.UserID, r.PathValue("id"), in.Name); err != nil {
		if errors.Is(err, ErrFileNotFound) {
			writeError(w, 404, "NOT_FOUND", "File not found.")
		} else {
			writeError(w, 400, "INVALID_INPUT", "Filename is invalid.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"updated": true})
}

func (h HTTPHandler) Move(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var in struct {
		FolderID string `json:"folderId"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in) != nil {
		writeError(w, 400, "INVALID_INPUT", "Invalid move payload.")
		return
	}
	if err := h.Service.Move(r.Context(), p.UserID, r.PathValue("id"), in.FolderID); err != nil {
		if errors.Is(err, ErrFileNotFound) {
			writeError(w, 404, "NOT_FOUND", "File or folder not found.")
		} else {
			writeError(w, 400, "INVALID_INPUT", "Folder is invalid.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"updated": true})
}

func (h HTTPHandler) Tags(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var in struct {
		Tags []string `json:"tags"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in) != nil {
		writeError(w, 400, "INVALID_INPUT", "Invalid tag payload.")
		return
	}
	if err := h.Service.SetTags(r.Context(), p.UserID, r.PathValue("id"), in.Tags); err != nil {
		if errors.Is(err, ErrFileNotFound) {
			writeError(w, 404, "NOT_FOUND", "File not found.")
		} else {
			writeError(w, 400, "INVALID_INPUT", "Tags are invalid.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"updated": true})
}

func (h HTTPHandler) CreatePublicShare(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var in struct {
		FileID    string     `json:"fileId"`
		FolderID  string     `json:"folderId"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in) != nil || (in.FileID == "" && in.FolderID == "") || (in.FileID != "" && in.FolderID != "") {
		writeError(w, 400, "INVALID_INPUT", "Exactly one file or folder ID is required.")
		return
	}
	var share PublicShare
	var err error
	if in.FolderID != "" {
		share, err = h.Service.CreatePublicFolderShare(r.Context(), p.UserID, in.FolderID, in.ExpiresAt)
	} else {
		share, err = h.Service.CreatePublicShare(r.Context(), p.UserID, in.FileID, in.ExpiresAt)
	}
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			writeError(w, 404, "NOT_FOUND", "File not found.")
		} else if errors.Is(err, ErrFolderNotFound) {
			writeError(w, 404, "NOT_FOUND", "Folder not found.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to create public share.")
		}
		return
	}
	writeJSON(w, 201, share)
}

func (h HTTPHandler) RevokePublicShare(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	if err := h.Service.RevokePublicShare(r.Context(), p.UserID, r.PathValue("id")); err != nil {
		if errors.Is(err, ErrShareNotFound) {
			writeError(w, 404, "NOT_FOUND", "Active share not found.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to revoke share.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"revoked": true})
}

func (h HTTPHandler) RevokePublicFolderShare(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	if err := h.Service.RevokePublicFolderShare(r.Context(), p.UserID, r.PathValue("id")); err != nil {
		if errors.Is(err, ErrShareNotFound) {
			writeError(w, 404, "NOT_FOUND", "Active share not found.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to revoke share.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"revoked": true})
}

func (h HTTPHandler) ShareAccess(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	items, err := h.Service.ListShareAccess(r.Context(), p.UserID, r.PathValue("id"), r.URL.Query().Get("folder") == "true")
	if err != nil {
		writeError(w, 500, "INTERNAL", "Unable to load sharing access.")
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (h HTTPHandler) RevokeDirectShare(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var err error
	if r.URL.Query().Get("folder") == "true" {
		err = h.Service.RevokeDirectFolderShare(r.Context(), p.UserID, r.PathValue("id"))
	} else {
		err = h.Service.RevokeDirectShare(r.Context(), p.UserID, r.PathValue("id"))
	}
	if err != nil {
		if errors.Is(err, ErrShareNotFound) {
			writeError(w, 404, "NOT_FOUND", "Share not found.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to revoke share.")
		}
		return
	}
	writeJSON(w, 200, map[string]bool{"revoked": true})
}

func (h HTTPHandler) CreateDirectShare(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, 401, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	var in struct {
		FileID         string `json:"fileId"`
		FolderID       string `json:"folderId"`
		RecipientEmail string `json:"recipientEmail"`
		Permission     string `json:"permission"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in) != nil {
		writeError(w, 400, "INVALID_INPUT", "Invalid share payload.")
		return
	}
	if (in.FileID == "") == (in.FolderID == "") {
		writeError(w, 400, "INVALID_INPUT", "Exactly one file or folder ID is required.")
		return
	}
	var err error
	if in.FolderID != "" {
		err = h.Service.CreateDirectFolderShare(r.Context(), p.UserID, in.FolderID, in.RecipientEmail, in.Permission)
	} else {
		err = h.Service.CreateDirectShare(r.Context(), p.UserID, in.FileID, in.RecipientEmail, in.Permission)
	}
	if err != nil {
		status := 400
		code := "INVALID_INPUT"
		msg := "Share request is invalid."
		if errors.Is(err, ErrFileNotFound) {
			status = 404
			code = "NOT_FOUND"
			msg = "File not found."
		}
		if errors.Is(err, ErrFolderNotFound) {
			status = 404
			code = "NOT_FOUND"
			msg = "Folder not found."
		}
		if errors.Is(err, ErrRecipientNotFound) {
			status = 404
			code = "NOT_FOUND"
			msg = "Recipient not found."
		}
		writeError(w, status, code, msg)
		return
	}
	writeJSON(w, 201, map[string]bool{"shared": true})
}

func (h HTTPHandler) Content(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	err := h.Service.ServeContent(r.Context(), w, r, p, r.PathValue("id"), r.URL.Query().Get("token"), false)
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			writeError(w, 404, "NOT_FOUND", "File not found.")
		} else if errors.Is(err, ErrContentForbidden) || errors.Is(err, policy.ErrUnauthenticated) {
			writeError(w, 403, "FORBIDDEN", "You are not authorized to access this file.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to download file.")
		}
	}
}

func (h HTTPHandler) Preview(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	err := h.Service.ServeContent(r.Context(), w, r, p, r.PathValue("id"), r.URL.Query().Get("token"), true)
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			writeError(w, 404, "NOT_FOUND", "File not found.")
		} else if errors.Is(err, ErrContentForbidden) || errors.Is(err, policy.ErrUnauthenticated) {
			writeError(w, 403, "FORBIDDEN", "You are not authorized to preview this file.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to preview file.")
		}
	}
}

func (h HTTPHandler) PublicContent(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	var err error
	if fileID := r.URL.Query().Get("fileId"); fileID != "" {
		err = h.Service.ServePublicFolderAsset(r.Context(), w, r, token, fileID, r.URL.Query().Get("preview") == "1")
	} else {
		err = h.Service.ServeContent(r.Context(), w, r, nil, "", token, r.URL.Query().Get("preview") == "1")
	}
	if err != nil {
		slog.Error("public content request failed", "path", r.URL.Path, "err", err)
		if errors.Is(err, ErrContentForbidden) {
			writeError(w, 404, "NOT_FOUND", "Public share not found or expired.")
		} else {
			writeError(w, 500, "INTERNAL", "Unable to download shared file.")
		}
	}
}

func (h HTTPHandler) PublicLanding(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	resource, err := h.Service.PublicResource(r.Context(), token)
	if err != nil {
		writeError(w, 404, "NOT_FOUND", "Public share not found or expired.")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(resource))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
