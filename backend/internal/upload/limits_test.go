package upload

import (
	"errors"
	"testing"
)

func TestNormalizeFilename(t *testing.T) {
	if got, err := NormalizeFilename(" report.pdf ", 255); err != nil || got != "report.pdf" {
		t.Fatalf("normalize valid filename: got %q err=%v", got, err)
	}
	for _, name := range []string{"", "..", `folder\\file.txt`, "bad\nname.txt"} {
		if _, err := NormalizeFilename(name, 255); !errors.Is(err, ErrInvalidFilename) {
			t.Fatalf("expected invalid filename for %q, got %v", name, err)
		}
	}
}

func TestValidateRequest(t *testing.T) {
	limits := Limits{MaxFileBytes: 10, MaxUploadBytes: 20, MaxFiles: 2, MaxFilenameSize: 255}
	if err := ValidateRequest(20, 2, limits); err != nil {
		t.Fatalf("boundary request rejected: %v", err)
	}
	if err := ValidateRequest(21, 1, limits); err == nil {
		t.Fatal("oversized request accepted")
	}
}
