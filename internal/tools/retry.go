package tools

import (
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	maxAttempts   = 3
	maxRetryAfter = 10 * time.Second
	maxDrainBytes = 4096
)

var baseDelay = 500 * time.Millisecond

type BuildRequest func(ctx context.Context) (*http.Request, error)

func Do(ctx context.Context, client *http.Client, build BuildRequest) (*http.Response, error) {
	for attempt := 1; ; attempt++ {
		req, err := build(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if !shouldRetry(resp, err) || attempt == maxAttempts || ctx.Err() != nil {
			return resp, err
		}

		wait := backoff(attempt, resp)
		discard(resp)
		if err := sleep(ctx, wait); err != nil {
			return nil, err
		}
	}
}

func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError
}

func backoff(attempt int, resp *http.Response) time.Duration {
	if after := retryAfter(resp); after > 0 {
		return min(after, maxRetryAfter)
	}
	exponential := baseDelay << (attempt - 1)
	jitter := time.Duration(rand.Int64N(int64(baseDelay)))
	return exponential + jitter
}

func retryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	seconds, err := strconv.Atoi(resp.Header.Get("Retry-After"))
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func discard(resp *http.Response) {
	if resp == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxDrainBytes))
	_ = resp.Body.Close()
}

func sleep(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
