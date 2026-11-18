package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oyamo/kwach/internal/application"
	"github.com/oyamo/kwach/internal/domain"
)

func TestStatusFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: domain.ErrProfileNotFound, want: http.StatusNotFound},
		{name: "unsafe script", err: domain.ErrUnsafeScript, want: http.StatusUnprocessableEntity},
		{name: "invalid profile", err: domain.ErrInvalidProfile, want: http.StatusBadRequest},
		{name: "invalid client id", err: domain.ErrInvalidClientID, want: http.StatusBadRequest},
		{name: "empty prompt", err: domain.ErrEmptyPrompt, want: http.StatusBadRequest},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFor(tt.err); got != tt.want {
				t.Errorf("statusFor() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInternalErrorsHideDetails(t *testing.T) {
	boom := errors.New("database exploded")
	ports := application.Ports{
		LoadProfile: func(context.Context, string) (domain.Profile, error) {
			return domain.Profile{}, boom
		},
	}
	logger := slog.New(slog.DiscardHandler)
	handler := NewHandler(application.NewRoaster(application.Narrator{}, ports, logger), logger).Routes()

	request := httptest.NewRequest(http.MethodGet, "/api/profiles/"+testClientID, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "internal error") {
		t.Errorf("body %q does not contain generic message", body)
	}
	if strings.Contains(body, "database exploded") {
		t.Errorf("body %q leaks the underlying error", body)
	}
}

func TestRoastIsRateLimitedPerClient(t *testing.T) {
	handler, _ := newTestRoutes(t)
	path := "/api/profiles/" + testClientID + "/roast"

	for range roastsPerClient {
		if code := send(t, handler, http.MethodPost, path, "").Code; code == http.StatusTooManyRequests {
			t.Fatal("request limited before reaching the quota")
		}
	}

	limited := send(t, handler, http.MethodPost, path, "")
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", limited.Code)
	}
	if limited.Header().Get("Retry-After") == "" {
		t.Error("429 response has no Retry-After header")
	}
}
