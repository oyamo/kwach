package infra

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func ConnectRedis(ctx context.Context, addr, password string, db int) (*Redis, error) {
	return connect(ctx, &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func ConnectRedisURL(ctx context.Context, rawURL string) (*Redis, error) {
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("infra: parse redis url: %w", err)
	}
	return connect(ctx, options)
}

func connect(ctx context.Context, options *redis.Options) (*Redis, error) {
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("infra: redis ping: %w", err)
	}
	return &Redis{client: client}, nil
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("infra: redis close: %w", err)
	}
	return nil
}
