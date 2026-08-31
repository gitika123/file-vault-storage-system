package files

import "testing"

func TestShareTokenIsOpaqueAndHashSized(t *testing.T) {
	token, hash, err := newShareToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(token) < 40 {
		t.Fatalf("token unexpectedly short: %d", len(token))
	}
	if len(hash) != 32 {
		t.Fatalf("hash length=%d", len(hash))
	}
	token2, _, _ := newShareToken()
	if token == token2 {
		t.Fatal("share tokens repeated")
	}
}
