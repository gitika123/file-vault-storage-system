package upload

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateMIMERejectsRenamedImage(t *testing.T) {
	if err := ValidateMIME("invoice.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "image/png"); err != ErrInvalidMIME {
		t.Fatalf("expected renamed image to be rejected, got %v", err)
	}
}

func TestValidateMIMEAllowsBroadMediaTypesWhenDetectorIsGeneric(t *testing.T) {
	for _, tc := range []struct{ name, mime string }{
		{"song.mp3", "audio/mpeg"},
		{"clip.mov", "video/quicktime"},
		{"photo.webp", "image/webp"},
		{"archive.7z", "application/x-7z-compressed"},
	} {
		if err := ValidateMIME(tc.name, tc.mime, "application/octet-stream"); err != nil {
			t.Fatalf("expected %s to be accepted: %v", tc.name, err)
		}
	}
}

func TestDetectFileAndValidatePDF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.bin")
	if err := os.WriteFile(path, []byte("%PDF-1.7\ncontent"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	detected, err := DetectFile(path)
	if err != nil || detected != "application/pdf" {
		t.Fatalf("expected PDF detection, got %q err=%v", detected, err)
	}
	if err := ValidateMIME("document.pdf", "application/pdf", detected); err != nil {
		t.Fatalf("valid PDF rejected: %v", err)
	}
}

func TestDetectOfficeDocumentFromZipStructure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.docx")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	archive := zip.NewWriter(file)
	for _, name := range []string{"[Content_Types].xml", "word/document.xml"} {
		entry, createErr := archive.Create(name)
		if createErr != nil {
			t.Fatalf("create zip entry: %v", createErr)
		}
		_, _ = entry.Write([]byte("fixture"))
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close fixture: %v", err)
	}
	detected, err := DetectFile(path)
	if err != nil || !strings.Contains(detected, "wordprocessingml.document") {
		t.Fatalf("expected DOCX detection, got %q err=%v", detected, err)
	}
}
