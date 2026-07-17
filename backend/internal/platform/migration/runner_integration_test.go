//go:build integration

package migration

import (
	"context"
	"os"
	"testing"

	"github.com/WeiGuang-2099/LaunchRail/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigrationsUpIsIdempotent(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	t.Cleanup(pool.Close)

	runner := New(pool, migrations.Files)
	if err := runner.Up(context.Background()); err != nil {
		t.Fatalf("first Up() error = %v", err)
	}
	if err := runner.Up(context.Background()); err != nil {
		t.Fatalf("second Up() error = %v", err)
	}

	var count int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM schema_migrations WHERE version = 1`).Scan(&count); err != nil {
		t.Fatalf("count schema migration error = %v", err)
	}
	if count != 1 {
		t.Fatalf("migration count = %d, want 1", count)
	}
}
