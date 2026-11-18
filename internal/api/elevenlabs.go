package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/oyamo/kwach/internal/domain"
	"github.com/oyamo/kwach/internal/tools"
)

const (
	elevenLabsBaseURL        = "https://api.elevenlabs.io"
	elevenLabsMaxAudioBytes  = 32 << 20
	elevenLabsRequestTimeout = 120 * time.Second
	elevenLabsAudioType      = "audio/mpeg"
)

var ErrElevenLabsAudioTooLarge = errors.New("elevenlabs: audio exceeds maximum size")

type ElevenLabsClient struct {
	apiKey       string
	voiceID      string
	modelID      string
	outputFormat string
	baseURL      string
	http         *http.Client
}

func NewElevenLabsClient(apiKey, voiceID, modelID, outputFormat string) *ElevenLabsClient {
	return &ElevenLabsClient{
		apiKey:       apiKey,
		voiceID:      voiceID,
		modelID:      modelID,
		outputFormat: outputFormat,
		baseURL:      elevenLabsBaseURL,
		http:         &http.Client{Timeout: elevenLabsRequestTimeout},
	}
}

type speechRequest struct {
	Text          string        `json:"text"`
	ModelID       string        `json:"model_id"`
	VoiceSettings voiceSettings `json:"voice_settings"`
}

type voiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
	Style           float64 `json:"style"`
	UseSpeakerBoost bool    `json:"use_speaker_boost"`
}

func (c *ElevenLabsClient) Synthesize(ctx context.Context, text string) (domain.Audio, error) {
	payload, err := json.Marshal(speechRequest{
		Text:    text,
		ModelID: c.modelID,
		VoiceSettings: voiceSettings{
			Stability:       0.45,
			SimilarityBoost: 0.8,
			Style:           0.2,
			UseSpeakerBoost: true,
		},
	})
	if err != nil {
		return domain.Audio{}, fmt.Errorf("elevenlabs: marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/text-to-speech/%s?output_format=%s",
		c.baseURL, url.PathEscape(c.voiceID), url.QueryEscape(c.outputFormat))

	resp, err := tools.Do(ctx, c.http, func(ctx context.Context) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("xi-api-key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", elevenLabsAudioType)
		return req, nil
	})
	if err != nil {
		return domain.Audio{}, fmt.Errorf("elevenlabs: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return domain.Audio{}, fmt.Errorf("elevenlabs: unexpected status %d: %s", resp.StatusCode, body)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, elevenLabsMaxAudioBytes+1))
	if err != nil {
		return domain.Audio{}, fmt.Errorf("elevenlabs: read audio: %w", err)
	}
	if len(data) > elevenLabsMaxAudioBytes {
		return domain.Audio{}, ErrElevenLabsAudioTooLarge
	}

	return domain.Audio{
		ContentType: elevenLabsContentType(resp.Header.Get("Content-Type")),
		Data:        data,
	}, nil
}

func elevenLabsContentType(value string) string {
	if value == "" {
		return elevenLabsAudioType
	}
	return value
}
