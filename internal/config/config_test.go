package config

import (
	"strings"
	"testing"
)

func TestLoadRequiresKeys(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ELEVENLABS_API_KEY", "")
	t.Setenv("ELEVENLABS_VOICE_IDS", "")
	t.Setenv("MONGODB_URI", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	for _, name := range []string{"OPENAI_API_KEY", "ELEVENLABS_API_KEY", "ELEVENLABS_VOICE_IDS", "MONGODB_URI"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("Load() error = %q, want it to mention %s", err, name)
		}
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("OPENAI_MODEL", "test-model")
	t.Setenv("ELEVENLABS_API_KEY", "eleven-key")
	t.Setenv("ELEVENLABS_VOICE_IDS", "voice-a, voice-b")
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("OPENAI_GUARD_MODEL", "")
	t.Setenv("ELEVENLABS_MODEL_ID", "")
	t.Setenv("ELEVENLABS_OUTPUT_FORMAT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.OpenAIBaseURL != defaultOpenAIBaseURL {
		t.Errorf("OpenAIBaseURL = %q, want %q", cfg.OpenAIBaseURL, defaultOpenAIBaseURL)
	}
	if cfg.OpenAIGuardModel != "test-model" {
		t.Errorf("OpenAIGuardModel = %q, want test-model", cfg.OpenAIGuardModel)
	}
	if cfg.ElevenLabsModelID != defaultElevenLabsModel {
		t.Errorf("ElevenLabsModelID = %q, want %q", cfg.ElevenLabsModelID, defaultElevenLabsModel)
	}
	if cfg.ElevenLabsFormat != defaultElevenLabsFormat {
		t.Errorf("ElevenLabsFormat = %q, want %q", cfg.ElevenLabsFormat, defaultElevenLabsFormat)
	}
}
