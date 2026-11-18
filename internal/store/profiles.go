package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/oyamo/kwach/internal/domain"
	"github.com/oyamo/kwach/internal/infra"
)

type ProfileStore struct {
	collection *mongo.Collection
}

type profileDocument struct {
	ClientID    string    `bson:"client_id"`
	Name        string    `bson:"name"`
	Category    string    `bson:"category"`
	Subcategory string    `bson:"subcategory"`
	Details     string    `bson:"details"`
	Intensity   string    `bson:"intensity"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func NewProfileStore(ctx context.Context, db *infra.MongoDB, collectionName string) (*ProfileStore, error) {
	collection := db.Collection(collectionName)
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "client_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := collection.Indexes().CreateOne(ctx, index); err != nil {
		return nil, fmt.Errorf("store: create index on %s: %w", collectionName, err)
	}
	return &ProfileStore{collection: collection}, nil
}

func (s *ProfileStore) StoreProfile(ctx context.Context, profile domain.Profile) error {
	document := profileDocument{
		ClientID:    profile.ClientID,
		Name:        profile.Name,
		Category:    string(profile.Category),
		Subcategory: profile.Subcategory,
		Details:     profile.Details,
		Intensity:   string(profile.Intensity),
		UpdatedAt:   time.Now().UTC(),
	}
	filter := bson.M{"client_id": profile.ClientID}
	if _, err := s.collection.ReplaceOne(ctx, filter, document, options.Replace().SetUpsert(true)); err != nil {
		return fmt.Errorf("store: upsert profile: %w", err)
	}
	return nil
}

func (s *ProfileStore) LoadProfile(ctx context.Context, clientID string) (domain.Profile, error) {
	var document profileDocument
	filter := bson.M{"client_id": clientID}
	err := s.collection.FindOne(ctx, filter).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	if err != nil {
		return domain.Profile{}, fmt.Errorf("store: find profile: %w", err)
	}

	return domain.Profile{
		ClientID:    document.ClientID,
		Name:        document.Name,
		Category:    domain.Category(document.Category),
		Subcategory: document.Subcategory,
		Details:     document.Details,
		Intensity:   domain.Intensity(document.Intensity),
	}, nil
}
