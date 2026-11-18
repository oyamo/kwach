package api

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/oyamo/kwach/internal/domain"
)

const (
	googleNewsBaseURL          = "https://news.google.com"
	googleNewsMaxResponseBytes = 1 << 20
	googleNewsRequestTimeout   = 10 * time.Second
	googleNewsCacheTTL         = 30 * time.Minute
	googleNewsMaxHeadlines     = 5
)

type GoogleNewsClient struct {
	baseURL string
	http    *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	headlines []domain.Headline
	expires   time.Time
}

func NewGoogleNewsClient() *GoogleNewsClient {
	return &GoogleNewsClient{
		baseURL: googleNewsBaseURL,
		http:    &http.Client{Timeout: googleNewsRequestTimeout},
		cache:   make(map[string]cacheEntry),
	}
}

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title  string `xml:"title"`
	Source string `xml:"source"`
}

func (c *GoogleNewsClient) Headlines(ctx context.Context, query string) ([]domain.Headline, error) {
	if cached, ok := c.cached(query); ok {
		return cached, nil
	}

	headlines, err := c.fetch(ctx, query)
	if err != nil {
		return nil, err
	}

	c.store(query, headlines)
	return headlines, nil
}

func (c *GoogleNewsClient) fetch(ctx context.Context, query string) ([]domain.Headline, error) {
	endpoint := c.baseURL + "/rss/search?" + url.Values{
		"q":    {query},
		"hl":   {"en-US"},
		"gl":   {"US"},
		"ceid": {"US:en"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("googlenews: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("googlenews: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("googlenews: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, googleNewsMaxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("googlenews: read feed: %w", err)
	}
	return parseHeadlines(body)
}

func parseHeadlines(body []byte) ([]domain.Headline, error) {
	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("googlenews: decode feed: %w", err)
	}

	headlines := make([]domain.Headline, 0, googleNewsMaxHeadlines)
	for _, item := range feed.Channel.Items {
		if len(headlines) == googleNewsMaxHeadlines {
			break
		}
		headlines = append(headlines, domain.Headline{Title: item.Title, Source: item.Source})
	}
	return headlines, nil
}

func (c *GoogleNewsClient) cached(query string) ([]domain.Headline, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.cache[query]
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}
	return entry.headlines, true
}

func (c *GoogleNewsClient) store(query string, headlines []domain.Headline) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[query] = cacheEntry{headlines: headlines, expires: time.Now().Add(googleNewsCacheTTL)}
}
