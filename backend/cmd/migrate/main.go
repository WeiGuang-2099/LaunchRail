package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/logging"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/migration"
	"github.com/WeiGuang-2099/LaunchRail/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := logging.New(os.Stdout, slog.LevelInfo)
	if err := run(context.Background(), os.Args[1:]); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info("database migration completed")
}

func run(ctx context.Context, arguments []string) error {
	if len(arguments) != 1 || (arguments[0] != "up" && arguments[0] != "down") {
		return fmt.Errorf("usage: migrate up|down")
	}

	databaseURL := strings.TrimSpace(os.Getenv("LAUNCHRAIL_DATABASE_URL"))
	if databaseURL == "" {
		return fmt.Errorf("LAUNCHRAIL_DATABASE_URL is required")
	}

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(connectCtx, databaseURL)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(connectCtx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	runner := migration.New(pool, migrations.Files)
	if arguments[0] == "up" {
		return runner.Up(ctx)
	}
	return runner.Down(ctx, 1)
}
