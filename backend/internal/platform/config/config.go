package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	defaultHTTPAddress     = ":8080"
	defaultLogLevel        = "info"
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	HTTPAddress     string
	DatabaseURL     string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("LAUNCHRAIL_DATABASE_URL"))

	httpAddress := strings.TrimSpace(os.Getenv("LAUNCHRAIL_HTTP_ADDR"))
	if httpAddress == "" {
		httpAddress = defaultHTTPAddress
	}

	logLevelValue := strings.TrimSpace(os.Getenv("LAUNCHRAIL_LOG_LEVEL"))
	if logLevelValue == "" {
		logLevelValue = defaultLogLevel
	}

	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(logLevelValue)); err != nil {
		return Config{}, fmt.Errorf("parse LAUNCHRAIL_LOG_LEVEL: %w", err)
	}

	shutdownTimeout := defaultShutdownTimeout
	if value := strings.TrimSpace(os.Getenv("LAUNCHRAIL_SHUTDOWN_TIMEOUT")); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse LAUNCHRAIL_SHUTDOWN_TIMEOUT: %w", err)
		}
		if parsed <= 0 {
			return Config{}, fmt.Errorf("LAUNCHRAIL_SHUTDOWN_TIMEOUT must be greater than zero")
		}
		shutdownTimeout = parsed
	}

	return Config{
		HTTPAddress:     httpAddress,
		DatabaseURL:     databaseURL,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}
