package application

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/oyamo/kwach/internal/domain"
)

func passModeration(context.Context, string) (bool, error) { return true, nil }

func TestNarrateChainsScriptIntoSpeech(t *testing.T) {
	generate := func(_ context.Context, prompt string) (string, error) {
		return "script for " + prompt, nil
	}
	synthesize := func(_ context.Context, _, text string) (domain.Audio, error) {
		return domain.Audio{ContentType: "audio/mpeg", Data: []byte(text)}, nil
	}

	narrator := NewNarrator(generate, passModeration, synthesize, slog.New(slog.DiscardHandler))
	narration, err := narrator.Narrate(t.Context(), "  cats  ", "voice-1")
	if err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}
	if got, want := narration.Script, "script for cats"; got != want {
		t.Errorf("Narrate() script = %q, want %q", got, want)
	}
	if got, want := string(narration.Audio.Data), "script for cats"; got != want {
		t.Errorf("Narrate() audio = %q, want %q", got, want)
	}
}

func TestNarrateStopsOnFailure(t *testing.T) {
	boom := errors.New("boom")
	okText := func(context.Context, string) (string, error) { return "fine script", nil }
	blockAll := func(context.Context, string) (bool, error) { return false, nil }
	failModeration := func(context.Context, string) (bool, error) { return false, boom }

	tests := []struct {
		name     string
		generate GenerateText
		moderate ModerateText
		prompt   string
		wantErr  error
	}{
		{
			name:     "empty prompt",
			generate: okText,
			moderate: passModeration,
			prompt:   "   ",
			wantErr:  domain.ErrEmptyPrompt,
		},
		{
			name:     "generation fails",
			generate: func(context.Context, string) (string, error) { return "", boom },
			moderate: passModeration,
			prompt:   "hi",
			wantErr:  boom,
		},
		{
			name:     "empty script",
			generate: func(context.Context, string) (string, error) { return "", nil },
			moderate: passModeration,
			prompt:   "hi",
			wantErr:  domain.ErrInvalidScript,
		},
		{
			name:     "script blocked by moderation",
			generate: okText,
			moderate: blockAll,
			prompt:   "hi",
			wantErr:  domain.ErrUnsafeScript,
		},
		{
			name:     "moderation fails closed",
			generate: okText,
			moderate: failModeration,
			prompt:   "hi",
			wantErr:  boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synthesizeCalled := false
			synthesize := func(context.Context, string, string) (domain.Audio, error) {
				synthesizeCalled = true
				return domain.Audio{}, nil
			}

			narrator := NewNarrator(tt.generate, tt.moderate, synthesize, slog.New(slog.DiscardHandler))
			_, err := narrator.Narrate(t.Context(), tt.prompt, "voice-1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Narrate() error = %v, want %v", err, tt.wantErr)
			}
			if synthesizeCalled {
				t.Error("synthesize was called after an earlier failure")
			}
		})
	}
}
