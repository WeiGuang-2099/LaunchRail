package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/application"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/domain"
	"github.com/google/uuid"
)

func TestCreateServiceRoute(t *testing.T) {
	service := mustDomainService(t, "payments-api")
	catalog := &catalogStub{register: func(_ context.Context, registration domain.Registration) (domain.Service, error) {
		if registration.ServiceKey != "payments-api" || registration.OwnerTeam != "payments-team" {
			t.Fatalf("registration = %#v", registration)
		}
		return service, nil
	}}
	recorder := performRequest(newTestRouter(catalog), http.MethodPost, "/api/v1/services", validCreateBody(), true)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Location") != "/api/v1/services/"+service.ID().String() {
		t.Fatalf("Location = %q", recorder.Header().Get("Location"))
	}
	var response serviceResponse
	decodeResponse(t, recorder, &response)
	if response.ID != service.ID() || response.ServiceKey != "payments-api" {
		t.Fatalf("response = %#v", response)
	}
}

func TestListServicesRouteUsesDefaults(t *testing.T) {
	service := mustDomainService(t, "catalog-api")
	var received application.ListOptions
	catalog := &catalogStub{list: func(_ context.Context, options application.ListOptions) (application.ServicePage, error) {
		received = options
		return application.ServicePage{Services: []domain.Service{service}, Total: 1}, nil
	}}
	recorder := performRequest(newTestRouter(catalog), http.MethodGet, "/api/v1/services", "", false)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if received.Status != application.ListStatusActive || received.Limit != 20 || received.Offset != 0 {
		t.Fatalf("options = %#v", received)
	}
	var response serviceListResponse
	decodeResponse(t, recorder, &response)
	if len(response.Items) != 1 || response.Pagination.Total != 1 {
		t.Fatalf("response = %#v", response)
	}
}

func TestServiceItemRoutes(t *testing.T) {
	service := mustDomainService(t, "orders-api")
	catalog := &catalogStub{
		get: func(_ context.Context, id uuid.UUID) (domain.Service, error) {
			if id != service.ID() {
				t.Fatalf("Get() id = %v", id)
			}
			return service, nil
		},
		update: func(_ context.Context, id uuid.UUID, metadata domain.MutableMetadata) (domain.Service, error) {
			if id != service.ID() || metadata.DisplayName != "Orders service" {
				t.Fatalf("Update() id = %v, metadata = %#v", id, metadata)
			}
			updated := service
			if err := updated.Update(metadata, service.UpdatedAt().Add(time.Minute)); err != nil {
				t.Fatalf("domain Update() error = %v", err)
			}
			return updated, nil
		},
		archive: func(_ context.Context, id uuid.UUID) error {
			if id != service.ID() {
				t.Fatalf("Archive() id = %v", id)
			}
			return nil
		},
	}
	router := newTestRouter(catalog)

	getRecorder := performRequest(router, http.MethodGet, "/api/v1/services/"+service.ID().String(), "", false)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, body = %s", getRecorder.Code, getRecorder.Body.String())
	}

	updateBody := `{"displayName":"Orders service","description":"Core orders","ownerTeam":"orders-team","repositoryUrl":"https://github.com/example/orders-api","lifecycle":"production"}`
	updateRecorder := performRequest(router, http.MethodPut, "/api/v1/services/"+service.ID().String(), updateBody, true)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, body = %s", updateRecorder.Code, updateRecorder.Body.String())
	}

	archiveRecorder := performRequest(router, http.MethodDelete, "/api/v1/services/"+service.ID().String(), "", false)
	if archiveRecorder.Code != http.StatusNoContent || archiveRecorder.Body.Len() != 0 {
		t.Fatalf("DELETE status = %d, body = %s", archiveRecorder.Code, archiveRecorder.Body.String())
	}
}

func TestServiceRoutesRejectInvalidRequests(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType bool
		wantStatus  int
	}{
		{name: "missing content type", method: http.MethodPost, path: "/api/v1/services", body: validCreateBody(), wantStatus: http.StatusUnsupportedMediaType},
		{name: "unknown field", method: http.MethodPost, path: "/api/v1/services", body: strings.TrimSuffix(validCreateBody(), "}") + `,"unknown":true}`, contentType: true, wantStatus: http.StatusBadRequest},
		{name: "multiple objects", method: http.MethodPost, path: "/api/v1/services", body: validCreateBody() + `{}`, contentType: true, wantStatus: http.StatusBadRequest},
		{name: "body too large", method: http.MethodPost, path: "/api/v1/services", body: `{"serviceKey":"` + strings.Repeat("a", maximumJSONBodyBytes) + `"}`, contentType: true, wantStatus: http.StatusBadRequest},
		{name: "invalid UUID", method: http.MethodGet, path: "/api/v1/services/not-a-uuid", wantStatus: http.StatusBadRequest},
		{name: "invalid status", method: http.MethodGet, path: "/api/v1/services?status=deleted", wantStatus: http.StatusBadRequest},
		{name: "invalid limit", method: http.MethodGet, path: "/api/v1/services?limit=101", wantStatus: http.StatusBadRequest},
		{name: "unsupported method", method: http.MethodPatch, path: "/api/v1/services", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown route", method: http.MethodGet, path: "/api/v1/unknown", wantStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := performRequest(newTestRouter(&catalogStub{}), test.method, test.path, test.body, test.contentType)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			var response errorResponse
			decodeResponse(t, recorder, &response)
			if response.Error.Code == "" || response.Error.RequestID == "" {
				t.Fatalf("error response = %#v", response)
			}
		})
	}
}

func TestServiceRouteMapsDomainAndRepositoryErrors(t *testing.T) {
	validation := &domain.ValidationErrors{Violations: []domain.FieldViolation{{Field: "displayName", Message: "is required"}}}
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "validation", err: validation, wantStatus: http.StatusUnprocessableEntity, wantCode: "validation_failed"},
		{name: "duplicate", err: application.ErrServiceKeyConflict, wantStatus: http.StatusConflict, wantCode: "service_key_conflict"},
		{name: "not found", err: application.ErrServiceNotFound, wantStatus: http.StatusNotFound, wantCode: "service_not_found"},
		{name: "archived", err: domain.ErrArchivedService, wantStatus: http.StatusConflict, wantCode: "service_archived"},
		{name: "internal", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := &catalogStub{register: func(context.Context, domain.Registration) (domain.Service, error) {
				return domain.Service{}, test.err
			}}
			recorder := performRequest(newTestRouter(catalog), http.MethodPost, "/api/v1/services", validCreateBody(), true)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			var response errorResponse
			decodeResponse(t, recorder, &response)
			if response.Error.Code != test.wantCode {
				t.Fatalf("error code = %q, want %q", response.Error.Code, test.wantCode)
			}
		})
	}
}

func newTestRouter(catalog ServiceCatalog) http.Handler {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return NewRouter(logger, func(context.Context) error { return nil }, catalog)
}

func performRequest(handler http.Handler, method, path, body string, contentType bool) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if contentType {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), destination); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %s", err, recorder.Body.String())
	}
}

func validCreateBody() string {
	return `{"serviceKey":"payments-api","displayName":"Payments API","description":"Core payments","ownerTeam":"payments-team","repositoryUrl":"https://github.com/example/payments-api","lifecycle":"development"}`
}

func mustDomainService(t *testing.T, key string) domain.Service {
	t.Helper()
	service, err := domain.NewService(domain.Registration{
		ServiceKey: key, DisplayName: "Test service", Description: "Test", OwnerTeam: "platform-team",
		RepositoryURL: "https://github.com/example/" + key, Lifecycle: domain.LifecycleDevelopment,
	}, time.Date(2026, time.July, 17, 1, 2, 3, 0, time.UTC))
	if err != nil {
		t.Fatalf("domain.NewService() error = %v", err)
	}
	return service
}

type catalogStub struct {
	register func(context.Context, domain.Registration) (domain.Service, error)
	get      func(context.Context, uuid.UUID) (domain.Service, error)
	list     func(context.Context, application.ListOptions) (application.ServicePage, error)
	update   func(context.Context, uuid.UUID, domain.MutableMetadata) (domain.Service, error)
	archive  func(context.Context, uuid.UUID) error
}

func (c *catalogStub) Register(ctx context.Context, registration domain.Registration) (domain.Service, error) {
	return c.register(ctx, registration)
}

func (c *catalogStub) Get(ctx context.Context, id uuid.UUID) (domain.Service, error) {
	return c.get(ctx, id)
}

func (c *catalogStub) List(ctx context.Context, options application.ListOptions) (application.ServicePage, error) {
	return c.list(ctx, options)
}

func (c *catalogStub) Update(ctx context.Context, id uuid.UUID, metadata domain.MutableMetadata) (domain.Service, error) {
	return c.update(ctx, id, metadata)
}

func (c *catalogStub) Archive(ctx context.Context, id uuid.UUID) error {
	return c.archive(ctx, id)
}
