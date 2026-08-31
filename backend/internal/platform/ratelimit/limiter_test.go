package ratelimit

import "testing"

func TestLimiterBurst(t *testing.T) {
	l := New(2, 2)
	if !l.Allow("alice") || !l.Allow("alice") || l.Allow("alice") {
		t.Fatal("burst was not enforced")
	}
	if !l.Allow("bob") {
		t.Fatal("rate limit leaked across users")
	}
}

func TestLimiterHonorsConfiguredBurst(t *testing.T) {
	l := New(2, 4)
	for i := 0; i < 4; i++ {
		if !l.Allow("alice") {
			t.Fatalf("request %d should fit in configured burst", i+1)
		}
	}
	if l.Allow("alice") {
		t.Fatal("fifth request should be rate limited")
	}
}
