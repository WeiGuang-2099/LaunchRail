package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/config"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/httpserver"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/logging"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/transport/httpapi"
)

func main() {
	logger := logging.New(os.Stdout, slog.LevelInfo)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration is invalid", "error", err)
		os.Exit(1)
	}
	logger = logging.New(os.Stdout, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("application stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	listener, err := net.Listen("tcp", cfg.HTTPAddress)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.HTTPAddress, err)
	}

	router := httpapi.NewRouter(logger, nil)
	server := httpserver.New(router, cfg.ShutdownTimeout)

	logger.Info("HTTP server started", "address", listener.Addr().String())
	if err := server.Serve(ctx, listener); err != nil {
		return err
	}
	logger.Info("HTTP server stopped")

	return nil
}
