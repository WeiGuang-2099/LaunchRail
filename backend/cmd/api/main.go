package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/application"
	catalogpostgres "github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/postgres"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/config"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/httpserver"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/logging"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/transport/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
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
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("LAUNCHRAIL_DATABASE_URL is required")
	}
	connectCtx, cancelConnect := context.WithTimeout(ctx, 10*time.Second)
	defer cancelConnect()
	pool, err := pgxpool.New(connectCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(connectCtx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	listener, err := net.Listen("tcp", cfg.HTTPAddress)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.HTTPAddress, err)
	}

	repository := catalogpostgres.NewServiceRepository(pool)
	catalog := application.NewCatalog(repository, time.Now)
	router := httpapi.NewRouter(logger, pool.Ping, catalog)
	server := httpserver.New(router, cfg.ShutdownTimeout)

	logger.Info("HTTP server started", "address", listener.Addr().String())
	if err := server.Serve(ctx, listener); err != nil {
		return err
	}
	logger.Info("HTTP server stopped")

	return nil
}
