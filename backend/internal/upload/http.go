package upload

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/balkanid/file-vault/internal/auth"
	"github.com/balkanid/file-vault/internal/storage"
)

type HTTPHandler struct {
	Service        Service
	MaxFileBytes   int64
	MaxUploadBytes int64
	MaxFiles       int
}

type resultResponse struct {
	Filename     string         `json:"filename"`
	FileID       string         `json:"fileId,omitempty"`
	SizeBytes    int64          `json:"sizeBytes,omitempty"`
	DetectedMIME string         `json:"detectedMime,omitempty"`
	Status       string         `json:"status"`
	Deduplicated bool           `json:"deduplicated,omitempty"`
	Error        *errorResponse `json:"error,omitempty"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h HTTPHandler) Upload(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	if r.ContentLength > h.MaxUploadBytes && r.ContentLength >= 0 {
		writeError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Upload request exceeds the configured limit.")
		return
	}
	if h.MaxFiles <= 0 || h.MaxFileBytes <= 0 || h.MaxUploadBytes < h.MaxFileBytes {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Upload limits are not configured.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxUploadBytes)
	multipart, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "A multipart upload is required.")
		return
	}
	results := make([]resultResponse, 0)
	fileCount := 0
	for {
		part, nextErr := multipart.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_INPUT", "Unable to read the upload.")
			return
		}
		if part.FileName() == "" {
			_, _ = io.Copy(io.Discard, part)
			_ = part.Close()
			continue
		}
		fileCount++
		if fileCount > h.MaxFiles {
			_ = part.Close()
			writeError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Upload contains too many files.")
			return
		}
		filename := part.FileName()
		result, uploadErr := h.Service.CreateFile(r.Context(), principal.UserID, filename, part.Header.Get("Content-Type"), part, h.MaxFileBytes)
		_ = part.Close()
		if uploadErr != nil {
			results = append(results, resultResponse{Filename: filename, Status: "rejected", Error: uploadError(uploadErr)})
			continue
		}
		results = append(results, resultResponse{Filename: result.Filename, FileID: result.FileID, SizeBytes: result.SizeBytes, DetectedMIME: result.DetectedMIME, Status: "created", Deduplicated: result.Deduplicated})
	}
	if fileCount == 0 {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "At least one file is required.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func uploadError(err error) *errorResponse {
	switch {
	case errors.Is(err, ErrQuotaExceeded):
		return &errorResponse{Code: "QUOTA_EXCEEDED", Message: "Upload would exceed the account storage quota."}
	case errors.Is(err, ErrInvalidMIME):
		return &errorResponse{Code: "INVALID_MIME", Message: "File content does not match its declared type."}
	case errors.Is(err, ErrInvalidFilename):
		return &errorResponse{Code: "INVALID_INPUT", Message: "Filename is invalid."}
	case errors.Is(err, ErrEmptyFile):
		return &errorResponse{Code: "INVALID_INPUT", Message: "Empty files cannot be uploaded."}
	case errors.Is(err, storage.ErrTooLarge):
		return &errorResponse{Code: "FILE_TOO_LARGE", Message: "File exceeds the configured size limit."}
	default:
		return &errorResponse{Code: "INTERNAL", Message: "File could not be stored."}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": errorResponse{Code: code, Message: message}})
}
