package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/config"
	"github.com/JoakimCarlsson/signal-ai-poc/internal/httpx"
	"github.com/JoakimCarlsson/signal-ai-poc/internal/otel"
	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
)

// flushTimeout bounds how long telemetry gets to drain on shutdown.
const flushTimeout = 5 * time.Second

// version is the reported build version; override with -ldflags.
var version = "dev"

// main runs the API and exits non-zero when it fails.
func main() {
	if err := run(); err != nil {
		slog.ErrorContext(context.Background(), "api failed", "error", err)
		os.Exit(1)
	}
}

// run wires the process together and serves until a signal arrives.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	shutdown, err := otel.Setup(ctx, otel.Config{
		ServiceName:          cfg.OTel.ServiceName,
		ServiceVersion:       version,
		OTLPExporterEndpoint: cfg.OTel.Endpoint,
		OTLPExporterHeaders:  cfg.OTel.Headers,
	})
	if err != nil {
		return fmt.Errorf("setting up telemetry: %w", err)
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(
			context.Background(),
			flushTimeout,
		)
		defer cancel()
		if err := shutdown(flushCtx); err != nil {
			slog.ErrorContext(flushCtx, "flushing telemetry", "error", err)
		}
	}()

	otel.SetupLogger(cfg.OTel.ServiceName, cfg.Log.Level, cfg.Log.Format)

	return httpx.New(
		httpx.Config{Addr: cfg.Addr},
		httpx.Deps{Shop: shop.NewStore()},
	).Start(ctx)
}
