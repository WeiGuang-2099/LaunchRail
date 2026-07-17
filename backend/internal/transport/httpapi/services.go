package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/application"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/domain"
	"github.com/google/uuid"
)

const maximumJSONBodyBytes = 1 << 20

type ServiceCatalog interface {
	Register(context.Context, domain.Registration) (domain.Service, error)
	Get(context.Context, uuid.UUID) (domain.Service, error)
	List(context.Context, application.ListOptions) (application.ServicePage, error)
	Update(context.Context, uuid.UUID, domain.MutableMetadata) (domain.Service, error)
	Archive(context.Context, uuid.UUID) error
}

type ServiceHandler struct {
	logger  *slog.Logger
	catalog ServiceCatalog
}

type serviceWriteRequest struct {
	DisplayName   string           `json:"displayName"`
	Description   string           `json:"description"`
	OwnerTeam     string           `json:"ownerTeam"`
	RepositoryURL string           `json:"repositoryUrl"`
	Lifecycle     domain.Lifecycle `json:"lifecycle"`
}

type serviceCreateRequest struct {
	ServiceKey string `json:"serviceKey"`
	serviceWriteRequest
}

type serviceResponse struct {
	ID            uuid.UUID        `json:"id"`
	ServiceKey    string           `json:"serviceKey"`
	DisplayName   string           `json:"displayName"`
	Description   string           `json:"description"`
	OwnerTeam     string           `json:"ownerTeam"`
	RepositoryURL string           `json:"repositoryUrl"`
	Lifecycle     domain.Lifecycle `json:"lifecycle"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
	ArchivedAt    *time.Time       `json:"archivedAt"`
}

type serviceListResponse struct {
	Items      []serviceResponse `json:"items"`
	Pagination pagination        `json:"pagination"`
}

type pagination struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	RequestID string        `json:"requestId"`
	Details   []errorDetail `json:"details"`
}

type errorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type requestError struct {
	status  int
	code    string
	message string
	details []errorDetail
}

func (e *requestError) Error() string { return e.message }

func NewServiceHandler(logger *slog.Logger, catalog ServiceCatalog) *ServiceHandler {
	return &ServiceHandler{logger: logger, catalog: catalog}
}

func (h *ServiceHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeAPIError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "The requested method is not supported.", nil)
	}
}

func (h *ServiceHandler) Item(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "invalid_request", "The service ID must be a valid UUID.", []errorDetail{{Field: "id", Message: "must be a valid UUID"}})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.archive(w, r, id)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeAPIError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "The requested method is not supported.", nil)
	}
}

func (h *ServiceHandler) create(w http.ResponseWriter, r *http.Request) {
	var request serviceCreateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, r, err)
		return
	}

	service, err := h.catalog.Register(r.Context(), domain.Registration{
		ServiceKey:    request.ServiceKey,
		DisplayName:   request.DisplayName,
		Description:   request.Description,
		OwnerTeam:     request.OwnerTeam,
		RepositoryURL: request.RepositoryURL,
		Lifecycle:     request.Lifecycle,
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/services/"+service.ID().String())
	writeJSON(w, http.StatusCreated, newServiceResponse(service))
}

func (h *ServiceHandler) list(w http.ResponseWriter, r *http.Request) {
	status := application.ListStatus(r.URL.Query().Get("status"))
	if status == "" {
		status = application.ListStatusActive
	}
	if !status.Valid() {
		writeAPIError(w, r, http.StatusBadRequest, "invalid_request", "The list query is invalid.", []errorDetail{{Field: "status", Message: "must be one of active, archived, or all"}})
		return
	}

	limit, err := parseIntegerQuery(r, "limit", 20, 1, 100)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	offset, err := parseIntegerQuery(r, "offset", 0, 0, int(^uint(0)>>1))
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	page, err := h.catalog.List(r.Context(), application.ListOptions{Status: status, Limit: limit, Offset: offset})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	items := make([]serviceResponse, 0, len(page.Services))
	for _, service := range page.Services {
		items = append(items, newServiceResponse(service))
	}
	writeJSON(w, http.StatusOK, serviceListResponse{
		Items: items,
		Pagination: pagination{
			Total: page.Total, Limit: limit, Offset: offset,
		},
	})
}

func (h *ServiceHandler) get(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	service, err := h.catalog.Get(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newServiceResponse(service))
}

func (h *ServiceHandler) update(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var request serviceWriteRequest
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, r, err)
		return
	}
	service, err := h.catalog.Update(r.Context(), id, domain.MutableMetadata{
		DisplayName:   request.DisplayName,
		Description:   request.Description,
		OwnerTeam:     request.OwnerTeam,
		RepositoryURL: request.RepositoryURL,
		Lifecycle:     request.Lifecycle,
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newServiceResponse(service))
}

func (h *ServiceHandler) archive(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.catalog.Archive(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ServiceHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var requestFailure *requestError
	if errors.As(err, &requestFailure) {
		writeAPIError(w, r, requestFailure.status, requestFailure.code, requestFailure.message, requestFailure.details)
		return
	}
	var validationErrors *domain.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make([]errorDetail, 0, len(validationErrors.Violations))
		for _, violation := range validationErrors.Violations {
			details = append(details, errorDetail{Field: violation.Field, Message: violation.Message})
		}
		writeAPIError(w, r, http.StatusUnprocessableEntity, "validation_failed", "One or more fields are invalid.", details)
		return
	}
	if errors.Is(err, application.ErrServiceNotFound) {
		writeAPIError(w, r, http.StatusNotFound, "service_not_found", "The requested service was not found.", nil)
		return
	}
	if errors.Is(err, application.ErrServiceKeyConflict) {
		writeAPIError(w, r, http.StatusConflict, "service_key_conflict", "A service with this service key already exists.", []errorDetail{{Field: "serviceKey", Message: "must be unique"}})
		return
	}
	if errors.Is(err, domain.ErrArchivedService) {
		writeAPIError(w, r, http.StatusConflict, "service_archived", "Archived services cannot be changed.", nil)
		return
	}

	h.logger.ErrorContext(r.Context(), "service API request failed", "request_id", RequestIDFromContext(r.Context()), "error", err)
	writeAPIError(w, r, http.StatusInternalServerError, "internal_error", "The server could not process the request.", nil)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &requestError{status: http.StatusUnsupportedMediaType, code: "unsupported_media_type", message: "Content-Type must be application/json."}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maximumJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		var maximumError *http.MaxBytesError
		message := "must contain one valid JSON object"
		if errors.Is(err, io.EOF) {
			message = "must not be empty"
		} else if errors.As(err, &maximumError) {
			message = fmt.Sprintf("must not exceed %d bytes", maximumJSONBodyBytes)
		}
		return &requestError{status: http.StatusBadRequest, code: "invalid_request", message: "The request body is invalid.", details: []errorDetail{{Field: "body", Message: message}}}
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return &requestError{status: http.StatusBadRequest, code: "invalid_request", message: "The request body is invalid.", details: []errorDetail{{Field: "body", Message: "must contain exactly one JSON object"}}}
	}
	return nil
}

func parseIntegerQuery(r *http.Request, name string, defaultValue, minimum, maximum int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, &requestError{
			status: http.StatusBadRequest, code: "invalid_request", message: "The list query is invalid.",
			details: []errorDetail{{Field: name, Message: fmt.Sprintf("must be an integer between %d and %d", minimum, maximum)}},
		}
	}
	return parsed, nil
}

func newServiceResponse(service domain.Service) serviceResponse {
	return serviceResponse{
		ID:            service.ID(),
		ServiceKey:    service.ServiceKey(),
		DisplayName:   service.DisplayName(),
		Description:   service.Description(),
		OwnerTeam:     service.OwnerTeam(),
		RepositoryURL: service.RepositoryURL(),
		Lifecycle:     service.Lifecycle(),
		CreatedAt:     service.CreatedAt(),
		UpdatedAt:     service.UpdatedAt(),
		ArchivedAt:    service.ArchivedAt(),
	}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, status int, code, message string, details []errorDetail) {
	if details == nil {
		details = []errorDetail{}
	}
	writeJSON(w, status, errorResponse{Error: apiError{
		Code: code, Message: message, RequestID: RequestIDFromContext(r.Context()), Details: details,
	}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
