package session

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSessionCreatesIndependentHashedTokens(t *testing.T) {
	first, err := New(time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("create first session: %v", err)
	}
	second, err := New(time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("create second session: %v", err)
	}
	if first.RawToken == second.RawToken || first.CSRFToken == second.CSRFToken {
		t.Fatal("session tokens must be unpredictable and independent")
	}
	if first.TokenHash == second.TokenHash || first.CSRFHash == second.CSRFHash {
		t.Fatal("token hashes must be independent")
	}
}

func TestValidateCSRF(t *testing.T) {
	newSession, err := New(time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	request := httptest.NewRequest("POST", "/api/auth/logout", nil)
	request.Header.Set(CSRFHeader, newSession.CSRFToken)
	if err := ValidateCSRF(request, newSession.CSRFHash); err != nil {
		t.Fatalf("valid csrf rejected: %v", err)
	}
	request.Header.Set(CSRFHeader, "wrong-token")
	if err := ValidateCSRF(request, newSession.CSRFHash); err == nil {
		t.Fatal("invalid csrf accepted")
	}
}
