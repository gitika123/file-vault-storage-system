package storage

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStageCommitOpenDelete(t *testing.T) {
	store := LocalStore{Root: t.TempDir()}
	object, err := store.Stage(context.Background(), strings.NewReader("hello vault"), 100)
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	if object.Size != int64(len("hello vault")) {
		t.Fatalf("unexpected size: %d", object.Size)
	}
	key := "ab-placeholder"
	if err := store.Commit(object, key); err != nil {
		t.Fatalf("commit: %v", err)
	}
	file, err := store.Open(key)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = file.Close()
	if err := store.Delete(key); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestStageRejectsOversizedStream(t *testing.T) {
	store := LocalStore{Root: t.TempDir()}
	_, err := store.Stage(context.Background(), strings.NewReader("0123456789"), 5)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestInvalidStorageKeyRejected(t *testing.T) {
	store := LocalStore{Root: t.TempDir()}
	if err := store.Commit(TempObject{Path: "unused"}, "../escape"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("expected invalid key, got %v", err)
	}
}
