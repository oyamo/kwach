package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSynthesizeSendsRequestAndReturnsAudio(t *testing.T) {
	audioBytes := []byte("fake-mp3-bytes")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/text-to-speech/voice-123" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("output_format"); got != "mp3_44100_128" {
			http.Error(w, "bad format: "+got, http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("xi-api-key"); got != "test-key" {
			http.Error(w, "bad key", http.StatusUnauthorized)
			return
		}

		var body speechRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body.Text != "hello world" || body.ModelID != "model-x" {
			http.Error(w, "unexpected body", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = io.Copy(w, bytes.NewReader(audioBytes))
	}))
	defer server.Close()

	client := NewElevenLabsClient("test-key", "voice-123", "model-x", "mp3_44100_128")
	client.baseURL = server.URL

	audio, err := client.Synthesize(t.Context(), "hello world")
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if audio.ContentType != "audio/mpeg" {
		t.Errorf("ContentType = %q, want audio/mpeg", audio.ContentType)
	}
	if !bytes.Equal(audio.Data, audioBytes) {
		t.Errorf("Data = %q, want %q", audio.Data, audioBytes)
	}
}

func TestSynthesizeNonOKStatusIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "quota exceeded", http.StatusPaymentRequired)
	}))
	defer server.Close()

	client := NewElevenLabsClient("test-key", "voice-123", "model-x", "mp3_44100_128")
	client.baseURL = server.URL

	if _, err := client.Synthesize(t.Context(), "hello"); err == nil {
		t.Fatal("Synthesize() error = nil, want error")
	}
}
