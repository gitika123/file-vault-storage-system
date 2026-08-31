package upload

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var ErrInvalidFilename = errors.New("invalid filename")

type Limits struct {
	MaxFileBytes    int64
	MaxUploadBytes  int64
	MaxFiles        int
	MaxFilenameSize int
}

func (l Limits) Validate() error {
	if l.MaxFileBytes <= 0 || l.MaxUploadBytes < l.MaxFileBytes || l.MaxFiles <= 0 || l.MaxFilenameSize <= 0 {
		return errors.New("invalid upload limits")
	}
	return nil
}

func NormalizeFilename(filename string, maxSize int) (string, error) {
	if maxSize <= 0 {
		return "", errors.New("filename limit must be positive")
	}
	filename = strings.TrimSpace(filename)
	if filename == "" || len([]rune(filename)) > maxSize || filename == "." || filename == ".." {
		return "", ErrInvalidFilename
	}
	if strings.ContainsAny(filename, `/\\`) {
		return "", ErrInvalidFilename
	}
	for _, char := range filename {
		if unicode.IsControl(char) {
			return "", ErrInvalidFilename
		}
	}
	return filename, nil
}

func ValidateRequest(contentLength int64, fileCount int, limits Limits) error {
	if err := limits.Validate(); err != nil {
		return err
	}
	if fileCount <= 0 || fileCount > limits.MaxFiles {
		return fmt.Errorf("upload must contain between 1 and %d files", limits.MaxFiles)
	}
	if contentLength > limits.MaxUploadBytes {
		return fmt.Errorf("upload request exceeds %d bytes", limits.MaxUploadBytes)
	}
	return nil
}
