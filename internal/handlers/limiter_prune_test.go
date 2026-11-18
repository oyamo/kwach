package handlers

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiterPrunesExpiredWindows(t *testing.T) {
	start := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	now := start
	limiter := newLimiter(1, time.Minute)
	limiter.now = func() time.Time { return now }

	for i := range maxTrackedKeys {
		limiter.allow(fmt.Sprintf("client-%d", i))
	}

	now = start.Add(2 * time.Minute)
	limiter.allow("fresh")

	if got := len(limiter.windows); got != 1 {
		t.Errorf("tracked keys = %d after pruning, want 1", got)
	}
}
