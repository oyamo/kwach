package handlers

import (
	"testing"
	"time"
)

func TestLimiterAllowsUpToLimitThenResets(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	limiter := newLimiter(2, time.Minute)
	limiter.now = func() time.Time { return now }

	results := []bool{
		limiter.allow("client-a"),
		limiter.allow("client-a"),
		limiter.allow("client-a"),
		limiter.allow("client-b"),
	}
	want := []bool{true, true, false, true}
	for i := range want {
		if results[i] != want[i] {
			t.Fatalf("call %d allowed = %v, want %v", i, results[i], want[i])
		}
	}

	now = now.Add(time.Minute)
	if !limiter.allow("client-a") {
		t.Error("allow() after window elapsed = false, want true")
	}
}
