package upload

import (
	"archive/zip"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidMIME = errors.New("content does not match declared MIME type")

var extensionMIMEs = map[string]string{
	// Images
	".pdf":  "application/pdf",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".tif":  "image/tiff",
	".tiff": "image/tiff",
	".svg":  "image/svg+xml",
	// Text and structured data
	".txt":  "text/plain",
	".csv":  "text/csv",
	".tsv":  "text/tab-separated-values",
	".md":   "text/markdown",
	".html": "text/html",
	".htm":  "text/html",
	".css":  "text/css",
	".js":   "text/javascript",
	".json": "application/json",
	".xml":  "application/xml",
	// Office and archive formats
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".doc":  "application/msword",
	".xls":  "application/vnd.ms-excel",
	".ppt":  "application/vnd.ms-powerpoint",
	".rtf":  "application/rtf",
	".zip":  "application/zip",
	".gz":   "application/gzip",
	".gzip": "application/gzip",
	".tar":  "application/x-tar",
	".7z":   "application/x-7z-compressed",
	".rar":  "application/vnd.rar",
	// Audio and video
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".oga":  "audio/ogg",
	".flac": "audio/flac",
	".m4a":  "audio/mp4",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
}

func DetectFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	if stat.Size() == 0 {
		return "application/octet-stream", nil
	}
	header := make([]byte, 512)
	read, err := io.ReadFull(file, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	}
	header = header[:read]
	if len(header) >= 12 && string(header[4:8]) == "ftyp" {
		return "video/mp4", nil
	}
	if len(header) >= 4 && string(header[:4]) == "PK\x03\x04" {
		if officeType, ok := detectOfficeZip(file, stat.Size()); ok {
			return officeType, nil
		}
	}
	return http.DetectContentType(header), nil
}

func ValidateMIME(filename, declared, detected string) error {
	declared = normalizeMIME(declared)
	if declared == "" {
		declared = "application/octet-stream"
	}
	detected = normalizeMIME(detected)
	expected, knownExtension := extensionMIMEs[strings.ToLower(filepath.Ext(filename))]
	if knownExtension {
		if declared != "application/octet-stream" && declared != expected {
			return ErrInvalidMIME
		}
		if detected != "application/octet-stream" && detected != expected {
			return ErrInvalidMIME
		}
		if expected == "application/pdf" && detected == "application/octet-stream" {
			return ErrInvalidMIME
		}
		return nil
	}
	if declared != "application/octet-stream" && detected != "application/octet-stream" && declared != detected {
		return ErrInvalidMIME
	}
	return nil
}

func normalizeMIME(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if mediaType, _, err := mime.ParseMediaType(value); err == nil {
		return mediaType
	}
	return strings.TrimSpace(strings.SplitN(value, ";", 2)[0])
}

func detectOfficeZip(file *os.File, size int64) (string, bool) {
	archive, err := zip.NewReader(file, size)
	if err != nil {
		return "", false
	}
	has := make(map[string]bool, len(archive.File))
	for _, entry := range archive.File {
		has[entry.Name] = true
	}
	switch {
	case has["word/document.xml"] && has["[Content_Types].xml"]:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true
	case has["xl/workbook.xml"] && has["[Content_Types].xml"]:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true
	case has["ppt/presentation.xml"] && has["[Content_Types].xml"]:
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation", true
	default:
		return "", false
	}
}
