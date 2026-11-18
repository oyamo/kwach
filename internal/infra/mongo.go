package infra

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

func ConnectMongo(ctx context.Context, uri, databaseName string, logger *slog.Logger) (*MongoDB, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("infra: mongo connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		disconnectQuietly(client)
		return nil, fmt.Errorf("infra: mongo ping: %w", err)
	}

	logger.Info("mongo connected", "database", databaseName)
	return &MongoDB{client: client, database: client.Database(databaseName)}, nil
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

func (m *MongoDB) Close(ctx context.Context) error {
	if err := m.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("infra: mongo disconnect: %w", err)
	}
	return nil
}

func disconnectQuietly(client *mongo.Client) {
	_ = client.Disconnect(context.Background())
}
