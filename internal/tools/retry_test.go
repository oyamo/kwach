package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func withFastRetries(t *testing.T) {
	t.Helper()
	original := baseDelay
	baseDelay = time.Millisecond
	t.Cleanup(func() { baseDelay = original })
}

func countingServer(t *testing.T, status func(hit int32) int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status(hits.Add(1)))
	}))
	t.Cleanup(server.Close)
	return server, &hits
}

func get(ctx context.Context, url string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func TestDoRetriesUntilSuccess(t *testing.T) {
	withFastRetries(t)
	server, hits := countingServer(t, func(hit int32) int {
		if hit < 2 {
			return http.StatusTooManyRequests
		}
		return http.StatusOK
	})

	resp, err := Do(t.Context(), server.Client(), func(ctx context.Context) (*http.Request, error) {
		return get(ctx, server.URL)
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if got := hits.Load(); got != 2 {
		t.Errorf("server hit %d times, want 2", got)
	}
}

func TestDoGivesUpAfterMaxAttempts(t *testing.T) {
	withFastRetries(t)
	server, hits := countingServer(t, func(int32) int { return http.StatusServiceUnavailable })

	resp, err := Do(t.Context(), server.Client(), func(ctx context.Context) (*http.Request, error) {
		return get(ctx, server.URL)
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want the last 503", resp.StatusCode)
	}
	if got := hits.Load(); got != maxAttempts {
		t.Errorf("server hit %d times, want %d", got, maxAttempts)
	}
}

func TestDoDoesNotRetryClientErrors(t *testing.T) {
	withFastRetries(t)
	server, hits := countingServer(t, func(int32) int { return http.StatusBadRequest })

	resp, err := Do(t.Context(), server.Client(), func(ctx context.Context) (*http.Request, error) {
		return get(ctx, server.URL)
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := hits.Load(); got != 1 {
		t.Errorf("server hit %d times, want 1", got)
	}
}
