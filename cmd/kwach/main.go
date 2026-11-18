package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/oyamo/kwach/internal/api"
	"github.com/oyamo/kwach/internal/application"
	"github.com/oyamo/kwach/internal/config"
	"github.com/oyamo/kwach/internal/handlers"
	"github.com/oyamo/kwach/internal/infra"
	"github.com/oyamo/kwach/internal/logging"
	"github.com/oyamo/kwach/internal/store"
)

const (
	shutdownTimeout    = 10 * time.Second
	profilesCollection = "profiles"
	historyRetention   = 7 * 24 * time.Hour
	redisDB            = 0
)

var errUsage = errors.New("usage: kwach <serve|narrate> [flags] [prompt]")

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "kwach:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if len(args) == 0 {
		return errUsage
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger, err := logging.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return err
	}

	switch args[0] {
	case "serve":
		return serve(ctx, cfg, logger)
	case "narrate":
		return narrate(ctx, cfg, logger, args[1:])
	default:
		return errUsage
	}
}

func serve(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	db, err := infra.ConnectMongo(ctx, cfg.MongoURI, cfg.MongoDatabase, logger)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			logger.Error("mongo disconnect failed", "error", err)
		}
	}()

	redisConn, err := infra.ConnectRedis(ctx, cfg.RedisAddr, cfg.RedisPassword, redisDB)
	if err != nil {
		return err
	}
	defer func() {
		if err := redisConn.Close(); err != nil {
			logger.Error("redis disconnect failed", "error", err)
		}
	}()

	profiles, err := store.NewProfileStore(ctx, db, profilesCollection)
	if err != nil {
		return err
	}
	history := store.NewHistoryStore(redisConn.Client(), historyRetention)

	ports := application.Ports{
		StoreProfile:   profiles.StoreProfile,
		LoadProfile:    profiles.LoadProfile,
		FetchHeadlines: api.NewGoogleNewsClient().Headlines,
		LoadHistory:    history.Load,
		AppendHistory:  history.Append,
		ClearHistory:   history.Clear,
	}
	roaster := application.NewRoaster(newNarrator(cfg, logger), ports, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handlers.NewHandler(roaster, logger).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      3 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()
	logger.Info("http server listening", "addr", cfg.HTTPAddr)

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func narrate(ctx context.Context, cfg config.Config, logger *slog.Logger, args []string) error {
	flags := flag.NewFlagSet("narrate", flag.ContinueOnError)
	outPath := flags.String("out", "kwach.mp3", "path of the audio file to write")
	if err := flags.Parse(args); err != nil {
		return err
	}

	narration, err := newNarrator(cfg, logger).Narrate(ctx, strings.Join(flags.Args(), " "))
	if err != nil {
		return err
	}
	if err := os.WriteFile(*outPath, narration.Audio.Data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", *outPath, err)
	}
	logger.Info("audio written", "path", *outPath, "bytes", len(narration.Audio.Data), "content_type", narration.Audio.ContentType)
	return nil
}

func newNarrator(cfg config.Config, logger *slog.Logger) application.Narrator {
	chat := api.NewGroqClient(cfg.GroqAPIKey, cfg.GroqModel)
	guard := api.NewGroqClient(cfg.GroqAPIKey, cfg.GroqGuardModel)
	voice := api.NewElevenLabsClient(cfg.ElevenLabsAPIKey, cfg.ElevenLabsVoiceID, cfg.ElevenLabsModelID, cfg.ElevenLabsFormat)
	return application.NewNarrator(chat.Complete, guard.Moderate, voice.Synthesize, logger)
}
