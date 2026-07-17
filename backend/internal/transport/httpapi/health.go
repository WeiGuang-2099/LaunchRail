package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
)

type ReadinessCheck func(context.Context) error

type HealthHandler struct {
	readinessCheck ReadinessCheck
}

func NewHealthHandler(readinessCheck ReadinessCheck) *HealthHandler {
	if readinessCheck == nil {
		readinessCheck = func(context.Context) error { return nil }
	}

	return &HealthHandler{readinessCheck: readinessCheck}
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	writeHealthStatus(w, http.StatusOK, "ok")
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.readinessCheck(r.Context()); err != nil {
		writeHealthStatus(w, http.StatusServiceUnavailable, "unavailable")
		return
	}

	writeHealthStatus(w, http.StatusOK, "ready")
}

func writeHealthStatus(w http.ResponseWriter, statusCode int, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{Status: status})
}
