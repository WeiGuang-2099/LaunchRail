package httpapi

import (
	"log/slog"
	"net/http"
)

func NewRouter(logger *slog.Logger, readinessCheck ReadinessCheck, catalog ServiceCatalog) http.Handler {
	healthHandler := NewHealthHandler(readinessCheck)
	serviceHandler := NewServiceHandler(logger, catalog)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler.Live)
	mux.HandleFunc("GET /readyz", healthHandler.Ready)
	mux.HandleFunc("/api/v1/services", serviceHandler.Collection)
	mux.HandleFunc("/api/v1/services/{id}", serviceHandler.Item)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(w, r, http.StatusNotFound, "route_not_found", "The requested API route was not found.", nil)
	})

	return RequestID(RequestLogger(logger)(mux))
}
