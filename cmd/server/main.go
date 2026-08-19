// Command server is the service entry point: load config, build the
// logger and telemetry, assemble the server, run it until an OS signal or
// a listener error, then shut down gracefully.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/config"
	"github.com/fauzie/golang-sekeleton/internal/server"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/telemetry"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New(logger.Config{
		Level:       cfg.Logging.Level,
		Environment: cfg.Server.Environment,
		Encoding:    cfg.Logging.Encoding,
	})
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	telemetryShutdown, err := telemetry.Init(ctx, telemetry.Config{
		Enabled:        cfg.Telemetry.Enabled,
		ServiceName:    cfg.Server.Name,
		ServiceVersion: "dev",
		Environment:    cfg.Server.Environment,
		OTLPEndpoint:   cfg.Telemetry.OTLPEndpoint,
		SampleRatio:    cfg.Telemetry.SampleRatio,
	})
	if err != nil {
		return fmt.Errorf("init telemetry: %w", err)
	}

	srv, err := server.New(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	errCh := srv.Start()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("server error, shutting down", zap.Error(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}

	shutdownTelemetryCtx, cancelTelemetry := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTelemetry()
	if err := telemetryShutdown(shutdownTelemetryCtx); err != nil {
		log.Error("telemetry shutdown error", zap.Error(err))
	}

	return nil
}
