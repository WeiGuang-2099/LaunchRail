package httpapi

import (
	"log/slog"
	"net/http"
)

func NewRouter(logger *slog.Logger, readinessCheck ReadinessCheck) http.Handler {
	healthHandler := NewHealthHandler(readinessCheck)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler.Live)
	mux.HandleFunc("GET /readyz", healthHandler.Ready)

	return RequestID(RequestLogger(logger)(mux))
}
