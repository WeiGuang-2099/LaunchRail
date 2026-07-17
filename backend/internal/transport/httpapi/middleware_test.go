package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/logging"
)

func TestRequestIDPreservesValidIncomingValue(t *testing.T) {
	const incomingRequestID = "client-request-123"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got != incomingRequestID {
			t.Fatalf("request ID in context = %q, want %q", got, incomingRequestID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(requestIDHeader, incomingRequestID)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(requestIDHeader); got != incomingRequestID {
		t.Fatalf("response request ID = %q, want %q", got, incomingRequestID)
	}
}

func TestRequestIDReplacesInvalidIncomingValue(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(requestIDHeader, "invalid request id")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	generated := recorder.Header().Get(requestIDHeader)
	if generated == "" || generated == "invalid request id" || !validRequestID(generated) {
		t.Fatalf("generated request ID = %q, want a new valid value", generated)
	}
}

func TestRequestLoggerWritesStructuredEvent(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(&output, slog.LevelInfo)
	handler := RequestID(RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/services", nil))

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if event["method"] != http.MethodPost {
		t.Fatalf("method = %v, want %s", event["method"], http.MethodPost)
	}
	if event["path"] != "/api/v1/services" {
		t.Fatalf("path = %v, want /api/v1/services", event["path"])
	}
	if event["status"] != float64(http.StatusAccepted) {
		t.Fatalf("status = %v, want %d", event["status"], http.StatusAccepted)
	}
	if event["request_id"] == "" {
		t.Fatal("request_id is empty")
	}
}
