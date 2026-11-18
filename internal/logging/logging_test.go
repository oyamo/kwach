package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewJSONLogsStructuredFields(t *testing.T) {
	var buffer bytes.Buffer
	logger, err := New(&buffer, "info", "json")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("hello", "key", "value")
	if !strings.Contains(buffer.String(), `"key":"value"`) {
		t.Errorf("output %q is not structured JSON", buffer.String())
	}
}

func TestNewFiltersBelowLevel(t *testing.T) {
	var buffer bytes.Buffer
	logger, err := New(&buffer, "warn", "text")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("quiet")
	if buffer.Len() != 0 {
		t.Errorf("info logged at warn level: %q", buffer.String())
	}
}

func TestNewRejectsBadSettings(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		format string
	}{
		{name: "bad level", level: "loud", format: "json"},
		{name: "bad format", level: "info", format: "xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(&bytes.Buffer{}, tt.level, tt.format); err == nil {
				t.Error("New() error = nil, want error")
			}
		})
	}
}
