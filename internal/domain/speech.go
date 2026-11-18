package domain

import (
	"errors"
	"strings"
)

const maxPromptLength = 2000

var (
	ErrEmptyPrompt   = errors.New("prompt is empty")
	ErrPromptTooLong = errors.New("prompt exceeds maximum length")
)

type Audio struct {
	ContentType string
	Data        []byte
}

func ValidatePrompt(prompt string) (string, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "", ErrEmptyPrompt
	}
	if len(trimmed) > maxPromptLength {
		return "", ErrPromptTooLong
	}
	return trimmed, nil
}
