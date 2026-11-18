package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const maxTrackedKeys = 10000

type limiter struct {
	mu      sync.Mutex
	limit   int
	period  time.Duration
	now     func() time.Time
	windows map[string]window
}

type window struct {
	start time.Time
	count int
}

func newLimiter(limit int, period time.Duration) *limiter {
	return &limiter{
		limit:   limit,
		period:  period,
		now:     time.Now,
		windows: make(map[string]window),
	}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if len(l.windows) >= maxTrackedKeys {
		l.prune(now)
	}

	current, ok := l.windows[key]
	if !ok || now.Sub(current.start) >= l.period {
		current = window{start: now}
	}
	if current.count >= l.limit {
		l.windows[key] = current
		return false
	}

	current.count++
	l.windows[key] = current
	return true
}

func (l *limiter) prune(now time.Time) {
	for key, current := range l.windows {
		if now.Sub(current.start) >= l.period {
			delete(l.windows, key)
		}
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
