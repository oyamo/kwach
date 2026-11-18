package logging

import (
	"fmt"
	"io"
	"log/slog"
)

func New(out io.Writer, level, format string) (*slog.Logger, error) {
	var threshold slog.Level
	if err := threshold.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("logging: invalid level %q: %w", level, err)
	}
	options := &slog.HandlerOptions{Level: threshold}

	switch format {
	case "json":
		return slog.New(slog.NewJSONHandler(out, options)), nil
	case "text":
		return slog.New(slog.NewTextHandler(out, options)), nil
	default:
		return nil, fmt.Errorf("logging: invalid format %q, want json or text", format)
	}
}
