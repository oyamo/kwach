package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oyamo/kwach/internal/application"
)

func TestCORSAllowsOnlyConfiguredOrigin(t *testing.T) {
	const allowed = "http://localhost:8080"
	handler := NewHandler(application.Roaster{}, slog.New(slog.DiscardHandler)).WithAllowedOrigin(allowed).Routes()

	tests := []struct {
		name       string
		method     string
		origin     string
		wantStatus int
		wantOrigin string
	}{
		{name: "preflight from allowed origin", method: http.MethodOptions, origin: allowed, wantStatus: http.StatusNoContent, wantOrigin: allowed},
		{name: "request from other origin", method: http.MethodOptions, origin: "http://evil.example", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/profiles/3f2b8c1e-9d4a-4b7e-8f21-0a1b2c3d4e5f", nil)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
		})
	}
}
