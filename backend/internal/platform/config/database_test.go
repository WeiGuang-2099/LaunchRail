package config

import "testing"

func TestLoadReadsDatabaseURL(t *testing.T) {
	t.Setenv("LAUNCHRAIL_DATABASE_URL", "postgres://launchrail:secret@db:5432/launchrail")
	t.Setenv("LAUNCHRAIL_HTTP_ADDR", "")
	t.Setenv("LAUNCHRAIL_LOG_LEVEL", "")
	t.Setenv("LAUNCHRAIL_SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabaseURL != "postgres://launchrail:secret@db:5432/launchrail" {
		t.Fatalf("DatabaseURL = %q, want configured URL", cfg.DatabaseURL)
	}
}
