package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>test</title>
<item><title>First story - Paper A</title><source url="https://a.example">Paper A</source></item>
<item><title>Second story - Paper B</title><source url="https://b.example">Paper B</source></item>
</channel></rss>`

func TestParseHeadlines(t *testing.T) {
	headlines, err := parseHeadlines([]byte(sampleFeed))
	if err != nil {
		t.Fatalf("parseHeadlines() error = %v", err)
	}
	if len(headlines) != 2 {
		t.Fatalf("got %d headlines, want 2", len(headlines))
	}
	if headlines[0].Title != "First story - Paper A" || headlines[0].Source != "Paper A" {
		t.Errorf("first headline = %+v", headlines[0])
	}
}

func TestParseHeadlinesCapsResults(t *testing.T) {
	var items strings.Builder
	for i := range 10 {
		fmt.Fprintf(&items, "<item><title>story %d</title><source>src</source></item>", i)
	}
	feed := "<rss><channel>" + items.String() + "</channel></rss>"

	headlines, err := parseHeadlines([]byte(feed))
	if err != nil {
		t.Fatalf("parseHeadlines() error = %v", err)
	}
	if len(headlines) != googleNewsMaxHeadlines {
		t.Errorf("got %d headlines, want %d", len(headlines), googleNewsMaxHeadlines)
	}
}

func TestHeadlinesUsesCache(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.URL.Query().Get("q"); got != "football" {
			http.Error(w, "unexpected query "+got, http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, sampleFeed)
	}))
	defer server.Close()

	client := NewGoogleNewsClient()
	client.baseURL = server.URL

	for range 2 {
		headlines, err := client.Headlines(t.Context(), "football")
		if err != nil {
			t.Fatalf("Headlines() error = %v", err)
		}
		if len(headlines) != 2 {
			t.Fatalf("got %d headlines, want 2", len(headlines))
		}
	}
	if requests != 1 {
		t.Errorf("server hit %d times, want 1 (second call served from cache)", requests)
	}
}

func TestHeadlinesNonOKStatusIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewGoogleNewsClient()
	client.baseURL = server.URL

	if _, err := client.Headlines(t.Context(), "football"); err == nil {
		t.Fatal("Headlines() error = nil, want error")
	}
}
