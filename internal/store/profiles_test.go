package store

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/oyamo/kwach/internal/domain"
	"github.com/oyamo/kwach/internal/infra"
)

// Runs only when MONGODB_TEST_URI is set, for example mongodb://localhost:27017.
func TestProfileStoreRoundTrip(t *testing.T) {
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		t.Skip("MONGODB_TEST_URI not set")
	}

	ctx := t.Context()
	db, err := infra.ConnectMongo(ctx, uri, "kwach_test", slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("ConnectMongo() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close(ctx) })

	profiles, err := NewProfileStore(ctx, db, "profiles_test")
	if err != nil {
		t.Fatalf("NewProfileStore() error = %v", err)
	}

	clientID := "7a1c2d3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e"
	profile := domain.Profile{
		ClientID:    clientID,
		Name:        "Alex",
		Category:    domain.CategoryMusic,
		Subcategory: "drummer",
		Details:     "always late",
		Intensity:   domain.IntensitySavage,
	}
	if err := profiles.StoreProfile(ctx, profile); err != nil {
		t.Fatalf("StoreProfile() error = %v", err)
	}

	got, err := profiles.LoadProfile(ctx, clientID)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}
	if got != profile {
		t.Errorf("LoadProfile() = %+v, want %+v", got, profile)
	}

	_, err = profiles.LoadProfile(ctx, "00000000-0000-4000-8000-000000000000")
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("LoadProfile() missing error = %v, want ErrProfileNotFound", err)
	}
}
