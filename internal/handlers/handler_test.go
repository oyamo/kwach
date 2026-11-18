package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oyamo/kwach/internal/application"
	"github.com/oyamo/kwach/internal/domain"
)

const testClientID = "3f2b8c1e-9d4a-4b7e-8f21-0a1b2c3d4e5f"

func newTestRoutes(t *testing.T) (http.Handler, map[string]domain.Profile) {
	t.Helper()
	store := map[string]domain.Profile{}
	history := map[string][]domain.Message{}
	logger := slog.New(slog.DiscardHandler)

	generate := func(context.Context, string) (string, error) { return "script", nil }
	passAll := func(context.Context, string) (bool, error) { return true, nil }
	synthesize := func(context.Context, string, string) (domain.Audio, error) {
		return domain.Audio{ContentType: "audio/mpeg", Data: []byte("mp3-bytes")}, nil
	}

	ports := application.Ports{
		VoiceIDs:       []string{"voice-a"},
		SaveVoiceIndex: func(context.Context, string, int) error { return nil },
		StoreProfile: func(_ context.Context, profile domain.Profile) error {
			store[profile.ClientID] = profile
			return nil
		},
		LoadProfile: func(_ context.Context, clientID string) (domain.Profile, error) {
			profile, ok := store[clientID]
			if !ok {
				return domain.Profile{}, domain.ErrProfileNotFound
			}
			return profile, nil
		},
		FetchHeadlines: func(context.Context, string) ([]domain.Headline, error) { return nil, nil },
		LoadHistory: func(_ context.Context, clientID string) ([]domain.Message, error) {
			return history[clientID], nil
		},
		AppendHistory: func(_ context.Context, clientID string, messages []domain.Message) error {
			history[clientID] = append(history[clientID], messages...)
			return nil
		},
		ClearHistory: func(_ context.Context, clientID string) error {
			delete(history, clientID)
			return nil
		},
	}

	narrator := application.NewNarrator(generate, passAll, synthesize, logger)
	roaster := application.NewRoaster(narrator, ports, logger)
	return NewHandler(roaster, logger).Routes(), store
}

func send(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestSaveAndGetProfile(t *testing.T) {
	handler, _ := newTestRoutes(t)
	path := "/api/profiles/" + testClientID
	payload := `{"name":"Alex","category":"sport","subcategory":"football","details":"","intensity":"savage"}`

	saved := send(t, handler, http.MethodPut, path, payload)
	if saved.Code != http.StatusNoContent {
		t.Fatalf("PUT status = %d, body %s", saved.Code, saved.Body)
	}

	loaded := send(t, handler, http.MethodGet, path, "")
	if loaded.Code != http.StatusOK {
		t.Fatalf("GET status = %d, body %s", loaded.Code, loaded.Body)
	}
	var output profileOutput
	if err := json.NewDecoder(loaded.Body).Decode(&output); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if output.Name != "Alex" || output.Intensity != "savage" {
		t.Errorf("GET body = %+v", output)
	}
}

func TestSaveRejectsBadInput(t *testing.T) {
	handler, _ := newTestRoutes(t)

	tests := []struct {
		name string
		path string
		body string
	}{
		{name: "bad client id", path: "/api/profiles/nope", body: `{"name":"Alex","category":"sport","subcategory":"x"}`},
		{name: "unknown field", path: "/api/profiles/" + testClientID, body: `{"name":"Alex","hacker":true}`},
		{name: "invalid profile", path: "/api/profiles/" + testClientID, body: `{"name":"","category":"sport","subcategory":"x"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := send(t, handler, http.MethodPut, tt.path, tt.body)
			if recorder.Code != http.StatusBadRequest {
				t.Errorf("PUT status = %d, want 400", recorder.Code)
			}
		})
	}
}

func TestGetMissingProfileIs404(t *testing.T) {
	handler, _ := newTestRoutes(t)
	recorder := send(t, handler, http.MethodGet, "/api/profiles/"+testClientID, "")
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET status = %d, want 404", recorder.Code)
	}
}

func TestRoastReturnsAudio(t *testing.T) {
	handler, store := newTestRoutes(t)
	store[testClientID] = domain.Profile{
		ClientID:    testClientID,
		Name:        "Alex",
		Category:    domain.CategoryCareer,
		Subcategory: "junior developer",
		Intensity:   domain.IntensityMedium,
	}

	recorder := send(t, handler, http.MethodPost, "/api/profiles/"+testClientID+"/roast", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("POST status = %d, body %s", recorder.Code, recorder.Body)
	}
	if got := recorder.Header().Get("Content-Type"); got != "audio/mpeg" {
		t.Errorf("Content-Type = %q, want audio/mpeg", got)
	}
	body, _ := io.ReadAll(recorder.Body)
	if !bytes.Equal(body, []byte("mp3-bytes")) {
		t.Errorf("body = %q, want mp3-bytes", body)
	}
}

func TestClearHistoryReturnsNoContent(t *testing.T) {
	handler, _ := newTestRoutes(t)

	recorder := send(t, handler, http.MethodDelete, "/api/profiles/"+testClientID+"/history", "")
	if recorder.Code != http.StatusNoContent {
		t.Errorf("DELETE status = %d, want 204", recorder.Code)
	}

	invalid := send(t, handler, http.MethodDelete, "/api/profiles/nope/history", "")
	if invalid.Code != http.StatusBadRequest {
		t.Errorf("DELETE bad id status = %d, want 400", invalid.Code)
	}
}
