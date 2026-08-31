package storage

import (
	"context"
	"io"
	"os"
)

// Store is the blob boundary used by upload and download services.
type Store interface {
	Initialize() error
	Stage(context.Context, io.Reader, int64) (TempObject, error)
	Commit(TempObject, string) error
	Open(string) (*os.File, error)
	Delete(string) error
	Discard(TempObject) error
}
