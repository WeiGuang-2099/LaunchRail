package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewServiceAcceptsValidMetadata(t *testing.T) {
	now := time.Date(2026, time.July, 17, 8, 30, 0, 0, time.UTC)
	service, err := NewService(validRegistration(), now)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if service.ID().String() == "" {
		t.Fatal("ID is empty")
	}
	if service.ServiceKey() != "payments-api" {
		t.Fatalf("ServiceKey() = %q, want payments-api", service.ServiceKey())
	}
	if service.CreatedAt() != now || service.UpdatedAt() != now {
		t.Fatalf("timestamps = %v/%v, want %v", service.CreatedAt(), service.UpdatedAt(), now)
	}
	if service.Archived() {
		t.Fatal("new service is archived")
	}
}

func TestNewServiceRejectsInvalidServiceKey(t *testing.T) {
	registration := validRegistration()
	registration.ServiceKey = "Payments API"

	_, err := NewService(registration, time.Now())
	assertFieldViolation(t, err, "serviceKey")
}

func TestNewServiceRejectsNonGitHubRepositoryURL(t *testing.T) {
	registration := validRegistration()
	registration.RepositoryURL = "https://gitlab.com/example/payments-api"

	_, err := NewService(registration, time.Now())
	assertFieldViolation(t, err, "repositoryUrl")
}

func TestUpdateServiceChangesMutableMetadataAndPreservesKey(t *testing.T) {
	createdAt := time.Date(2026, time.July, 17, 8, 30, 0, 0, time.UTC)
	service, err := NewService(validRegistration(), createdAt)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	updatedAt := createdAt.Add(time.Minute)
	err = service.Update(MutableMetadata{
		DisplayName:   "Payments Service",
		Description:   "Updated description",
		OwnerTeam:     "commerce-platform",
		RepositoryURL: "https://github.com/example/payments-service",
		Lifecycle:     LifecycleProduction,
	}, updatedAt)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if service.ServiceKey() != "payments-api" {
		t.Fatalf("ServiceKey() = %q, want immutable payments-api", service.ServiceKey())
	}
	if service.DisplayName() != "Payments Service" || service.UpdatedAt() != updatedAt {
		t.Fatalf("service was not updated: name=%q updatedAt=%v", service.DisplayName(), service.UpdatedAt())
	}
}

func TestUpdateServiceRejectsArchivedService(t *testing.T) {
	service, err := NewService(validRegistration(), time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if err := service.Archive(time.Now()); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	err = service.Update(MutableMetadata{
		DisplayName:   "New name",
		OwnerTeam:     "team",
		RepositoryURL: "https://github.com/example/new-name",
		Lifecycle:     LifecycleDevelopment,
	}, time.Now().Add(time.Minute))
	if !errors.Is(err, ErrArchivedService) {
		t.Fatalf("Update() error = %v, want ErrArchivedService", err)
	}
}

func TestArchiveServiceIsIdempotent(t *testing.T) {
	createdAt := time.Date(2026, time.July, 17, 8, 30, 0, 0, time.UTC)
	service, err := NewService(validRegistration(), createdAt)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	archivedAt := createdAt.Add(time.Minute)
	if err := service.Archive(archivedAt); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}
	if err := service.Archive(archivedAt.Add(time.Minute)); err != nil {
		t.Fatalf("second Archive() error = %v", err)
	}

	if got := service.ArchivedAt(); got == nil || !got.Equal(archivedAt) {
		t.Fatalf("ArchivedAt() = %v, want %v", got, archivedAt)
	}
}

func validRegistration() Registration {
	return Registration{
		ServiceKey:    "payments-api",
		DisplayName:   "Payments API",
		Description:   "Processes payments",
		OwnerTeam:     "payments-platform",
		RepositoryURL: "https://github.com/example/payments-api",
		Lifecycle:     LifecycleDevelopment,
	}
}

func assertFieldViolation(t *testing.T, err error, field string) {
	t.Helper()

	var validationErrors *ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("error = %v, want ValidationErrors", err)
	}
	for _, violation := range validationErrors.Violations {
		if violation.Field == field {
			return
		}
	}
	t.Fatalf("error = %v, want violation for %s", err, field)
}
