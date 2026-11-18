package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	defaultOpenAIBaseURL    = "https://api.openai.com/v1"
	defaultElevenLabsModel  = "eleven_multilingual_v2"
	defaultElevenLabsFormat = "mp3_44100_128"
	defaultMongoDatabase    = "kwach"
	defaultHTTPAddr         = ":8080"
	defaultLogLevel         = "info"
	defaultLogFormat        = "json"
	defaultRedisAddr        = "localhost:6379"
)

type Config struct {
	OpenAIBaseURL      string
	OpenAIAPIKey       string
	OpenAIModel        string
	OpenAIGuardModel   string
	ElevenLabsAPIKey   string
	ElevenLabsVoiceIDs []string
	ElevenLabsModelID  string
	ElevenLabsFormat   string
	MongoURI           string
	MongoDatabase      string
	HTTPAddr           string
	CORSOrigin         string
	LogLevel           string
	LogFormat          string
	RedisAddr          string
	RedisPassword      string
	RedisURL           string
}

func Load() (Config, error) {
	cfg := Config{
		OpenAIBaseURL:      envOr("OPENAI_BASE_URL", defaultOpenAIBaseURL),
		OpenAIAPIKey:       os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:        os.Getenv("OPENAI_MODEL"),
		OpenAIGuardModel:   envOr("OPENAI_GUARD_MODEL", os.Getenv("OPENAI_MODEL")),
		ElevenLabsAPIKey:   os.Getenv("ELEVENLABS_API_KEY"),
		ElevenLabsVoiceIDs: parseList(os.Getenv("ELEVENLABS_VOICE_IDS")),
		ElevenLabsModelID:  envOr("ELEVENLABS_MODEL_ID", defaultElevenLabsModel),
		ElevenLabsFormat:   envOr("ELEVENLABS_OUTPUT_FORMAT", defaultElevenLabsFormat),
		MongoURI:           os.Getenv("MONGODB_URI"),
		MongoDatabase:      envOr("MONGODB_DATABASE", defaultMongoDatabase),
		HTTPAddr:           envOr("HTTP_ADDR", defaultHTTPAddr),
		CORSOrigin:         os.Getenv("CORS_ORIGIN"),
		LogLevel:           envOr("LOG_LEVEL", defaultLogLevel),
		LogFormat:          envOr("LOG_FORMAT", defaultLogFormat),
		RedisAddr:          envOr("REDIS_ADDR", defaultRedisAddr),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisURL:           os.Getenv("REDIS_URL"),
	}

	required := []struct {
		name  string
		value string
	}{
		{"OPENAI_API_KEY", cfg.OpenAIAPIKey},
		{"OPENAI_MODEL", cfg.OpenAIModel},
		{"ELEVENLABS_API_KEY", cfg.ElevenLabsAPIKey},
		{"ELEVENLABS_VOICE_IDS", strings.Join(cfg.ElevenLabsVoiceIDs, ",")},
		{"MONGODB_URI", cfg.MongoURI},
	}

	var missing []error
	for _, field := range required {
		if field.value == "" {
			missing = append(missing, fmt.Errorf("%s is not set", field.name))
		}
	}
	if err := errors.Join(missing...); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

func parseList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func envOr(name, fallback string) string {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return fallback
	}
	return value
}
