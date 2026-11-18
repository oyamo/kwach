package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/oyamo/kwach/internal/domain"
)

const historyKeyPrefix = "chat:history:"

type HistoryStore struct {
	client    *redis.Client
	retention time.Duration
}

type historyRecord struct {
	Role    domain.Role `json:"role"`
	Content string      `json:"content"`
	At      time.Time   `json:"at"`
}

func NewHistoryStore(client *redis.Client, retention time.Duration) *HistoryStore {
	return &HistoryStore{client: client, retention: retention}
}

func (s *HistoryStore) Append(ctx context.Context, clientID string, messages []domain.Message) error {
	values := make([]any, 0, len(messages))
	for _, message := range messages {
		encoded, err := json.Marshal(historyRecord{Role: message.Role, Content: message.Content, At: message.At})
		if err != nil {
			return fmt.Errorf("store: encode message: %w", err)
		}
		values = append(values, encoded)
	}

	key := historyKey(clientID)
	_, err := s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.RPush(ctx, key, values...)
		pipe.LTrim(ctx, key, -domain.MaxHistoryMessages, -1)
		pipe.Expire(ctx, key, s.retention)
		return nil
	})
	if err != nil {
		return fmt.Errorf("store: append history: %w", err)
	}
	return nil
}

func (s *HistoryStore) Load(ctx context.Context, clientID string) ([]domain.Message, error) {
	values, err := s.client.LRange(ctx, historyKey(clientID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("store: load history: %w", err)
	}

	messages := make([]domain.Message, 0, len(values))
	for _, value := range values {
		var record historyRecord
		if err := json.Unmarshal([]byte(value), &record); err != nil {
			return nil, fmt.Errorf("store: decode message: %w", err)
		}
		messages = append(messages, domain.Message{Role: record.Role, Content: record.Content, At: record.At})
	}
	return messages, nil
}

func (s *HistoryStore) Clear(ctx context.Context, clientID string) error {
	if err := s.client.Del(ctx, historyKey(clientID)).Err(); err != nil {
		return fmt.Errorf("store: clear history: %w", err)
	}
	return nil
}

func historyKey(clientID string) string {
	return historyKeyPrefix + clientID
}
