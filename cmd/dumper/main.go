package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/nerdneilsfield/dumper/internal/api"
	"github.com/nerdneilsfield/dumper/internal/bot"
	"github.com/nerdneilsfield/dumper/internal/config"
	"github.com/nerdneilsfield/dumper/internal/ingest"
	"github.com/nerdneilsfield/dumper/internal/llm"
	"github.com/nerdneilsfield/dumper/internal/store"
)

func main() {
	level := slog.LevelInfo
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Set log level from config
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	// Initialize store manager
	stores, err := store.NewManager(cfg.DataDir)
	if err != nil {
		return fmt.Errorf("create store manager: %w", err)
	}
	defer stores.Close()

	// Initialize LLM client
	llmClient := llm.NewClient(cfg.OpenRouterKey, cfg.OpenRouterModel)

	// Initialize processing pipeline
	pipeline := ingest.NewPipeline(llmClient, stores)

	// Initialize bot
	tgBot, err := bot.New(cfg.TelegramToken, pipeline, stores, cfg.WebAppURL)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	// Initialize API server
	apiServer := api.NewServer(stores, cfg.TelegramToken, llmClient)

	// Setup graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// Run bot
	g.Go(func() error {
		slog.Info("starting telegram bot")
		return tgBot.Run(ctx)
	})

	// Run HTTP server
	g.Go(func() error {
		addr := fmt.Sprintf(":%d", cfg.HTTPPort)
		slog.Info("starting http server", "addr", addr)

		server := &http.Server{Addr: addr, Handler: apiServer}

		go func() {
			<-ctx.Done()
			server.Shutdown(context.Background())
		}()

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	return g.Wait()
}
