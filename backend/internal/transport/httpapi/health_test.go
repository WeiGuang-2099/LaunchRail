package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerLive(t *testing.T) {
	handler := NewHealthHandler(nil)
	recorder := httptest.NewRecorder()

	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assertHealthResponse(t, recorder, http.StatusOK, "ok")
}

func TestHealthHandlerReady(t *testing.T) {
	handler := NewHealthHandler(func(context.Context) error { return nil })
	recorder := httptest.NewRecorder()

	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	assertHealthResponse(t, recorder, http.StatusOK, "ready")
}

func TestHealthHandlerReadyWhenDependencyIsUnavailable(t *testing.T) {
	handler := NewHealthHandler(func(context.Context) error { return errors.New("dependency unavailable") })
	recorder := httptest.NewRecorder()

	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	assertHealthResponse(t, recorder, http.StatusServiceUnavailable, "unavailable")
}

func assertHealthResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantCode int, wantStatus string) {
	t.Helper()

	if recorder.Code != wantCode {
		t.Fatalf("status code = %d, want %d", recorder.Code, wantCode)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", recorder.Header().Get("Content-Type"))
	}

	var response struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Status != wantStatus {
		t.Fatalf("status = %q, want %q", response.Status, wantStatus)
	}
}
