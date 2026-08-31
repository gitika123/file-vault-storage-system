package files

import (
	"testing"
	"time"
)

func TestCursorRoundTrip(t *testing.T) {
	now := time.Date(2026, 8, 28, 12, 30, 0, 123, time.UTC)
	encoded := encodeCursor(now, "file-id")
	got, id, err := decodeCursor(encoded)
	if err != nil || !got.Equal(now) || id != "file-id" {
		t.Fatalf("cursor round trip: %v %v %v", got, id, err)
	}
}
func TestCursorRejectsInvalidInput(t *testing.T) {
	if _, _, err := decodeCursor("not-a-cursor"); err != ErrInvalidCursor {
		t.Fatalf("expected invalid cursor, got %v", err)
	}
}
