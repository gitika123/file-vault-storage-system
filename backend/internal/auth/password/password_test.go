package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	encoded, err := Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	ok, err := Verify(encoded, "correct horse battery staple")
	if err != nil || !ok {
		t.Fatalf("expected password to verify, ok=%v err=%v", ok, err)
	}
	wrong, err := Verify(encoded, "incorrect password value")
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}
	if wrong {
		t.Fatal("wrong password verified")
	}
}

func TestHashRejectsShortPassword(t *testing.T) {
	if _, err := Hash("short"); err == nil {
		t.Fatal("expected short password to be rejected")
	}
}
