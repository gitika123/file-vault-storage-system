package storage

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidKey = errors.New("invalid blob storage key")
	ErrTooLarge   = errors.New("stream exceeds configured size limit")
)

type TempObject struct {
	Path   string
	Size   int64
	SHA256 [32]byte
}

type LocalStore struct {
	Root string
}

func (s LocalStore) Initialize() error {
	if s.Root == "" {
		return errors.New("blob store root is required")
	}
	return os.MkdirAll(filepath.Join(s.Root, ".tmp"), 0o750)
}

func (s LocalStore) Stage(ctx context.Context, source io.Reader, maxBytes int64) (TempObject, error) {
	if maxBytes <= 0 {
		return TempObject{}, errors.New("maxBytes must be positive")
	}
	if err := s.Initialize(); err != nil {
		return TempObject{}, fmt.Errorf("initialize blob store: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Join(s.Root, ".tmp"), "upload-*")
	if err != nil {
		return TempObject{}, fmt.Errorf("create temporary upload: %w", err)
	}
	tempPath := temp.Name()
	keep := false
	defer func() {
		_ = temp.Close()
		if !keep {
			_ = os.Remove(tempPath)
		}
	}()

	hasher := sha256.New()
	reader := &contextReader{ctx: ctx, reader: io.LimitReader(source, maxBytes+1)}
	written, err := io.Copy(io.MultiWriter(temp, hasher), reader)
	if err != nil {
		return TempObject{}, fmt.Errorf("stage upload: %w", err)
	}
	if written > maxBytes {
		return TempObject{}, ErrTooLarge
	}
	if err := temp.Sync(); err != nil {
		return TempObject{}, fmt.Errorf("sync temporary upload: %w", err)
	}
	if err := temp.Close(); err != nil {
		return TempObject{}, fmt.Errorf("close temporary upload: %w", err)
	}
	keep = true
	var digest [32]byte
	copy(digest[:], hasher.Sum(nil))
	return TempObject{Path: tempPath, Size: written, SHA256: digest}, nil
}

func (s LocalStore) Commit(temp TempObject, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if temp.Path == "" {
		return errors.New("temporary object path is required")
	}
	if err := s.Initialize(); err != nil {
		return fmt.Errorf("initialize blob store: %w", err)
	}
	destination := filepath.Join(s.Root, key)
	if err := os.MkdirAll(filepath.Dir(destination), 0o750); err != nil {
		return fmt.Errorf("create blob directory: %w", err)
	}
	if err := os.Rename(temp.Path, destination); err != nil {
		return fmt.Errorf("commit blob: %w", err)
	}
	return nil
}

func (s LocalStore) Open(key string) (*os.File, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	return os.Open(filepath.Join(s.Root, key))
}

func (s LocalStore) Delete(key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	return os.Remove(filepath.Join(s.Root, key))
}

func (s LocalStore) Discard(temp TempObject) error {
	if temp.Path == "" {
		return nil
	}
	return os.Remove(temp.Path)
}

func validateKey(key string) error {
	if key == "" || filepath.IsAbs(key) || strings.ContainsAny(key, `\\/`) || key == "." || key == ".." || strings.Contains(key, "..") {
		return ErrInvalidKey
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(p)
	}
}
