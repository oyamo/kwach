package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/oyamo/kwach/internal/domain"
)

type GenerateText func(ctx context.Context, prompt string) (string, error)

type ModerateText func(ctx context.Context, text string) (bool, error)

type SynthesizeSpeech func(ctx context.Context, voiceID, text string) (domain.Audio, error)

type Narrator struct {
	generateText     GenerateText
	moderateText     ModerateText
	synthesizeSpeech SynthesizeSpeech
	logger           *slog.Logger
}

func NewNarrator(
	generateText GenerateText,
	moderateText ModerateText,
	synthesizeSpeech SynthesizeSpeech,
	logger *slog.Logger,
) Narrator {
	return Narrator{
		generateText:     generateText,
		moderateText:     moderateText,
		synthesizeSpeech: synthesizeSpeech,
		logger:           logger,
	}
}

func (n Narrator) Narrate(ctx context.Context, prompt, voiceID string) (domain.Narration, error) {
	validPrompt, err := domain.ValidatePrompt(prompt)
	if err != nil {
		return domain.Narration{}, fmt.Errorf("narrate: %w", err)
	}

	scriptStarted := time.Now()
	script, err := n.generateText(ctx, validPrompt)
	if err != nil {
		return domain.Narration{}, fmt.Errorf("narrate: generate text: %w", err)
	}
	if err := domain.ValidateScript(script); err != nil {
		return domain.Narration{}, fmt.Errorf("narrate: %w", err)
	}
	n.logger.Info("script generated", "characters", len(script), "duration", time.Since(scriptStarted))

	safe, err := n.moderateText(ctx, script)
	if err != nil {
		return domain.Narration{}, fmt.Errorf("narrate: moderate: %w", err)
	}
	if !safe {
		n.logger.Warn("script blocked by moderation", "characters", len(script))
		return domain.Narration{}, fmt.Errorf("narrate: %w", domain.ErrUnsafeScript)
	}

	speechStarted := time.Now()
	audio, err := n.synthesizeSpeech(ctx, voiceID, script)
	if err != nil {
		return domain.Narration{}, fmt.Errorf("narrate: synthesize speech: %w", err)
	}
	n.logger.Info("speech synthesized", "bytes", len(audio.Data), "duration", time.Since(speechStarted))
	return domain.Narration{Script: script, Audio: audio}, nil
}
