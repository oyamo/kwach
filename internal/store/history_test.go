package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/oyamo/kwach/internal/domain"
	"github.com/oyamo/kwach/internal/infra"
)

// Runs only when REDIS_TEST_ADDR is set, for example localhost:6379.
func TestHistoryStoreKeepsLatestMessages(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR not set")
	}

	ctx := t.Context()
	conn, err := infra.ConnectRedis(ctx, addr, "", 0)
	if err != nil {
		t.Fatalf("ConnectRedis() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	clientID := "7a1c2d3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e"
	history := NewHistoryStore(conn.Client(), time.Minute)
	t.Cleanup(func() { _ = history.Clear(context.Background(), clientID) })

	if err := history.Clear(ctx, clientID); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	var batch []domain.Message
	for i := range domain.MaxHistoryMessages + 2 {
		batch = append(batch, domain.Message{
			Role:    domain.RoleAssistant,
			Content: fmt.Sprintf("roast %d", i),
			At:      time.Now().UTC(),
		})
	}
	if err := history.Append(ctx, clientID, batch); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	got, err := history.Load(ctx, clientID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got) != domain.MaxHistoryMessages {
		t.Fatalf("Load() returned %d messages, want %d", len(got), domain.MaxHistoryMessages)
	}
	if got[0].Content != "roast 2" {
		t.Errorf("oldest kept = %q, want roast 2", got[0].Content)
	}
	if got[len(got)-1].Content != "roast 11" {
		t.Errorf("newest = %q, want roast 11", got[len(got)-1].Content)
	}
}
