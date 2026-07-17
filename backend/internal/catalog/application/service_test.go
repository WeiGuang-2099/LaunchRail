package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/domain"
	"github.com/google/uuid"
)

func TestCatalogRegisterUsesClockAndPersistsService(t *testing.T) {
	now := time.Date(2026, time.July, 17, 3, 4, 5, 0, time.UTC)
	var persisted domain.Service
	repository := &repositoryStub{create: func(_ context.Context, service domain.Service) error {
		persisted = service
		return nil
	}}
	catalog := NewCatalog(repository, func() time.Time { return now })

	service, err := catalog.Register(context.Background(), validRegistration("catalog-api"))
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if persisted.ID() != service.ID() || !service.CreatedAt().Equal(now) {
		t.Fatalf("persisted service = %#v, returned service = %#v", persisted, service)
	}
}

func TestCatalogRegisterReturnsValidationBeforeRepository(t *testing.T) {
	called := false
	repository := &repositoryStub{create: func(context.Context, domain.Service) error {
		called = true
		return nil
	}}
	catalog := NewCatalog(repository, time.Now)

	_, err := catalog.Register(context.Background(), domain.Registration{})
	var validationErrors *domain.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("Register() error = %v, want ValidationErrors", err)
	}
	if called {
		t.Fatal("repository Create() was called for invalid registration")
	}
}

func TestCatalogRegisterPreservesDuplicateConflict(t *testing.T) {
	repository := &repositoryStub{create: func(context.Context, domain.Service) error {
		return ErrServiceKeyConflict
	}}
	catalog := NewCatalog(repository, time.Now)

	_, err := catalog.Register(context.Background(), validRegistration("duplicate"))
	if !errors.Is(err, ErrServiceKeyConflict) {
		t.Fatalf("Register() error = %v, want ErrServiceKeyConflict", err)
	}
}

func TestCatalogUpdateLoadsChangesAndPersistsService(t *testing.T) {
	createdAt := time.Date(2026, time.July, 17, 1, 0, 0, 0, time.UTC)
	now := createdAt.Add(time.Hour)
	service := mustService(t, "update-service", createdAt)
	var persisted domain.Service
	repository := &repositoryStub{
		getByID: func(context.Context, uuid.UUID) (domain.Service, error) { return service, nil },
		update: func(_ context.Context, service domain.Service) error {
			persisted = service
			return nil
		},
	}
	catalog := NewCatalog(repository, func() time.Time { return now })

	updated, err := catalog.Update(context.Background(), service.ID(), domain.MutableMetadata{
		DisplayName: "Updated service", Description: "Updated", OwnerTeam: "runtime-team",
		RepositoryURL: "https://github.com/example/updated-service", Lifecycle: domain.LifecycleProduction,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.DisplayName() != "Updated service" || persisted.ID() != service.ID() || !updated.UpdatedAt().Equal(now) {
		t.Fatalf("updated service = %#v, persisted service = %#v", updated, persisted)
	}
}

func TestCatalogArchiveIsIdempotent(t *testing.T) {
	createdAt := time.Date(2026, time.July, 17, 1, 0, 0, 0, time.UTC)
	now := createdAt.Add(time.Hour)
	service := mustService(t, "archive-service", createdAt)
	var archivedAt time.Time
	repository := &repositoryStub{
		getByID: func(context.Context, uuid.UUID) (domain.Service, error) { return service, nil },
		archive: func(_ context.Context, _ uuid.UUID, value time.Time) error {
			archivedAt = value
			return nil
		},
	}
	catalog := NewCatalog(repository, func() time.Time { return now })

	if err := catalog.Archive(context.Background(), service.ID()); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}
	if !archivedAt.Equal(now) {
		t.Fatalf("archivedAt = %v, want %v", archivedAt, now)
	}
}

func TestCatalogGetPreservesNotFound(t *testing.T) {
	repository := &repositoryStub{getByID: func(context.Context, uuid.UUID) (domain.Service, error) {
		return domain.Service{}, ErrServiceNotFound
	}}
	catalog := NewCatalog(repository, time.Now)

	_, err := catalog.Get(context.Background(), uuid.New())
	if !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("Get() error = %v, want ErrServiceNotFound", err)
	}
}

func validRegistration(key string) domain.Registration {
	return domain.Registration{
		ServiceKey: key, DisplayName: "Catalog API", Description: "Test service", OwnerTeam: "platform-team",
		RepositoryURL: "https://github.com/example/" + key, Lifecycle: domain.LifecycleDevelopment,
	}
}

func mustService(t *testing.T, key string, now time.Time) domain.Service {
	t.Helper()
	service, err := domain.NewService(validRegistration(key), now)
	if err != nil {
		t.Fatalf("domain.NewService() error = %v", err)
	}
	return service
}

type repositoryStub struct {
	create  func(context.Context, domain.Service) error
	getByID func(context.Context, uuid.UUID) (domain.Service, error)
	list    func(context.Context, ListOptions) (ServicePage, error)
	update  func(context.Context, domain.Service) error
	archive func(context.Context, uuid.UUID, time.Time) error
}

func (r *repositoryStub) Create(ctx context.Context, service domain.Service) error {
	return r.create(ctx, service)
}

func (r *repositoryStub) GetByID(ctx context.Context, id uuid.UUID) (domain.Service, error) {
	return r.getByID(ctx, id)
}

func (r *repositoryStub) List(ctx context.Context, options ListOptions) (ServicePage, error) {
	return r.list(ctx, options)
}

func (r *repositoryStub) Update(ctx context.Context, service domain.Service) error {
	return r.update(ctx, service)
}

func (r *repositoryStub) Archive(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.archive(ctx, id, at)
}
