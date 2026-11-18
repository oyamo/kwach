package main

import (
	"errors"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_MODEL", "test-model")
	t.Setenv("ELEVENLABS_API_KEY", "eleven-test-key")
	t.Setenv("ELEVENLABS_VOICE_IDS", "voice-test")
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
}

func TestRunRequiresCommand(t *testing.T) {
	if err := run(nil); !errors.Is(err, errUsage) {
		t.Errorf("run(nil) error = %v, want errUsage", err)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	setRequiredEnv(t)

	if err := run([]string{"bogus"}); !errors.Is(err, errUsage) {
		t.Errorf("run(bogus) error = %v, want errUsage", err)
	}
}

func TestRunFailsWithoutConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ELEVENLABS_API_KEY", "")
	t.Setenv("ELEVENLABS_VOICE_IDS", "")
	t.Setenv("MONGODB_URI", "")

	if err := run([]string{"serve"}); err == nil {
		t.Error("run(serve) error = nil, want config error")
	}
}
