package api

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveContent(content string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"`+content+`"}}]}`)
	}))
}

func TestCompleteReturnsTrimmedContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			http.Error(w, "bad auth: "+got, http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"  Hello there.  "}}]}`)
	}))
	defer server.Close()

	client := NewGroqClient("test-key", "test-model")
	client.baseURL = server.URL

	got, err := client.Complete(t.Context(), "say hello")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if got != "Hello there." {
		t.Errorf("Complete() = %q, want %q", got, "Hello there.")
	}
}

func TestCompleteErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{name: "non-200 status", status: http.StatusBadRequest, body: "bad request"},
		{name: "no choices", status: http.StatusOK, body: `{"choices":[]}`, wantErr: ErrGroqNoContent},
		{name: "blank content", status: http.StatusOK, body: `{"choices":[{"message":{"content":"   "}}]}`, wantErr: ErrGroqNoContent},
		{name: "invalid json", status: http.StatusOK, body: `not json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer server.Close()

			client := NewGroqClient("test-key", "test-model")
			client.baseURL = server.URL

			_, err := client.Complete(t.Context(), "prompt")
			if err == nil {
				t.Fatal("Complete() error = nil, want error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Complete() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestModerateVerdicts(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantSafe bool
		wantErr  error
	}{
		{name: "safe", content: "safe", wantSafe: true},
		{name: "unsafe with category", content: "unsafe\\nS1", wantSafe: false},
		{name: "unexpected answer", content: "maybe", wantErr: ErrGroqUnexpectedVerdict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := serveContent(tt.content)
			defer server.Close()

			client := NewGroqClient("test-key", "guard-model")
			client.baseURL = server.URL

			safe, err := client.Moderate(t.Context(), "some script")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Moderate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Moderate() error = %v", err)
			}
			if safe != tt.wantSafe {
				t.Errorf("Moderate() = %v, want %v", safe, tt.wantSafe)
			}
		})
	}
}
