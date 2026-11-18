package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		want    string
		wantErr error
	}{
		{name: "trims whitespace", prompt: "  hello  ", want: "hello"},
		{name: "blank is empty", prompt: " \t\n ", wantErr: ErrEmptyPrompt},
		{name: "too long", prompt: strings.Repeat("a", maxPromptLength+1), wantErr: ErrPromptTooLong},
		{name: "at max length", prompt: strings.Repeat("a", maxPromptLength), want: strings.Repeat("a", maxPromptLength)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidatePrompt(tt.prompt)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidatePrompt() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ValidatePrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}
