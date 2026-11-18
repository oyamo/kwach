package config

import (
	"errors"
	"fmt"
	"os"
)

const (
	defaultGroqModel        = "llama-3.3-70b-versatile"
	defaultGroqGuardModel   = "meta-llama/llama-guard-4-12b"
	defaultElevenLabsModel  = "eleven_multilingual_v2"
	defaultElevenLabsFormat = "mp3_44100_128"
	defaultMongoDatabase    = "kwach"
	defaultHTTPAddr         = ":8080"
	defaultLogLevel         = "info"
	defaultLogFormat        = "json"
	defaultRedisAddr        = "localhost:6379"
)

type Config struct {
	GroqAPIKey        string
	GroqModel         string
	GroqGuardModel    string
	ElevenLabsAPIKey  string
	ElevenLabsVoiceID string
	ElevenLabsModelID string
	ElevenLabsFormat  string
	MongoURI          string
	MongoDatabase     string
	HTTPAddr          string
	LogLevel          string
	LogFormat         string
	RedisAddr         string
	RedisPassword     string
}

func Load() (Config, error) {
	cfg := Config{
		GroqAPIKey:        os.Getenv("GROQ_API_KEY"),
		GroqModel:         envOr("GROQ_MODEL", defaultGroqModel),
		GroqGuardModel:    envOr("GROQ_GUARD_MODEL", defaultGroqGuardModel),
		ElevenLabsAPIKey:  os.Getenv("ELEVENLABS_API_KEY"),
		ElevenLabsVoiceID: os.Getenv("ELEVENLABS_VOICE_ID"),
		ElevenLabsModelID: envOr("ELEVENLABS_MODEL_ID", defaultElevenLabsModel),
		ElevenLabsFormat:  envOr("ELEVENLABS_OUTPUT_FORMAT", defaultElevenLabsFormat),
		MongoURI:          os.Getenv("MONGODB_URI"),
		MongoDatabase:     envOr("MONGODB_DATABASE", defaultMongoDatabase),
		HTTPAddr:          envOr("HTTP_ADDR", defaultHTTPAddr),
		LogLevel:          envOr("LOG_LEVEL", defaultLogLevel),
		LogFormat:         envOr("LOG_FORMAT", defaultLogFormat),
		RedisAddr:         envOr("REDIS_ADDR", defaultRedisAddr),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
	}

	required := []struct {
		name  string
		value string
	}{
		{"GROQ_API_KEY", cfg.GroqAPIKey},
		{"ELEVENLABS_API_KEY", cfg.ElevenLabsAPIKey},
		{"ELEVENLABS_VOICE_ID", cfg.ElevenLabsVoiceID},
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

func envOr(name, fallback string) string {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return fallback
	}
	return value
}
