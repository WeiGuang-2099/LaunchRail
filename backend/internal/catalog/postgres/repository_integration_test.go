//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/application"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/domain"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/platform/migration"
	"github.com/WeiGuang-2099/LaunchRail/backend/migrations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresServiceRepositoryCreateThenGetPersistsAllFields(t *testing.T) {
	repository := setupRepository(t)
	service := newTestService(t, "catalog-create-get")

	if err := repository.Create(context.Background(), service); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	stored, err := repository.GetByID(context.Background(), service.ID())
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if stored.ID() != service.ID() || stored.ServiceKey() != service.ServiceKey() ||
		stored.DisplayName() != service.DisplayName() || stored.Description() != service.Description() ||
		stored.OwnerTeam() != service.OwnerTeam() || stored.RepositoryURL() != service.RepositoryURL() ||
		stored.Lifecycle() != service.Lifecycle() || !stored.CreatedAt().Equal(service.CreatedAt()) ||
		!stored.UpdatedAt().Equal(service.UpdatedAt()) || stored.Archived() {
		t.Fatalf("stored service = %#v, want all fields from created service", stored)
	}
}

func TestPostgresServiceRepositoryReturnsDuplicateKeyConflict(t *testing.T) {
	repository := setupRepository(t)
	first := newTestService(t, "catalog-duplicate")
	second := newTestService(t, "catalog-duplicate")

	if err := repository.Create(context.Background(), first); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if err := repository.Create(context.Background(), second); !errors.Is(err, application.ErrServiceKeyConflict) {
		t.Fatalf("second Create() error = %v, want ErrServiceKeyConflict", err)
	}
}

func TestPostgresServiceRepositoryListFiltersArchivedServices(t *testing.T) {
	repository := setupRepository(t)
	active := newTestService(t, "catalog-active")
	archived := newTestService(t, "catalog-archived")
	if err := repository.Create(context.Background(), active); err != nil {
		t.Fatalf("create active service: %v", err)
	}
	if err := repository.Create(context.Background(), archived); err != nil {
		t.Fatalf("create archived service: %v", err)
	}
	if err := repository.Archive(context.Background(), archived.ID(), time.Now().UTC()); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	activePage, err := repository.List(context.Background(), application.ListOptions{Status: application.ListStatusActive, Limit: 100})
	if err != nil {
		t.Fatalf("List(active) error = %v", err)
	}
	archivedPage, err := repository.List(context.Background(), application.ListOptions{Status: application.ListStatusArchived, Limit: 100})
	if err != nil {
		t.Fatalf("List(archived) error = %v", err)
	}

	if !containsService(activePage.Services, active.ID()) || containsService(activePage.Services, archived.ID()) {
		t.Fatalf("active services = %#v, want only active test service", activePage.Services)
	}
	if !containsService(archivedPage.Services, archived.ID()) || containsService(archivedPage.Services, active.ID()) {
		t.Fatalf("archived services = %#v, want only archived test service", archivedPage.Services)
	}
}

func TestPostgresServiceRepositoryUpdatesMutableMetadata(t *testing.T) {
	repository := setupRepository(t)
	service := newTestService(t, "catalog-update")
	if err := repository.Create(context.Background(), service); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err := service.Update(domain.MutableMetadata{
		DisplayName:   "Updated service",
		Description:   "Updated description",
		OwnerTeam:     "updated-team",
		RepositoryURL: "https://github.com/example/updated-service",
		Lifecycle:     domain.LifecycleProduction,
	}, service.UpdatedAt().Add(time.Minute))
	if err != nil {
		t.Fatalf("domain Update() error = %v", err)
	}
	if err := repository.Update(context.Background(), service); err != nil {
		t.Fatalf("repository Update() error = %v", err)
	}

	stored, err := repository.GetByID(context.Background(), service.ID())
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if stored.DisplayName() != "Updated service" || stored.OwnerTeam() != "updated-team" || stored.Lifecycle() != domain.LifecycleProduction {
		t.Fatalf("stored metadata was not updated: %#v", stored)
	}
}

func TestPostgresServiceRepositoryArchiveIsIdempotent(t *testing.T) {
	repository := setupRepository(t)
	service := newTestService(t, "catalog-archive")
	if err := repository.Create(context.Background(), service); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	archivedAt := service.UpdatedAt().Add(time.Minute)
	if err := repository.Archive(context.Background(), service.ID(), archivedAt); err != nil {
		t.Fatalf("first Archive() error = %v", err)
	}
	if err := repository.Archive(context.Background(), service.ID(), archivedAt.Add(time.Minute)); err != nil {
		t.Fatalf("second Archive() error = %v", err)
	}

	stored, err := repository.GetByID(context.Background(), service.ID())
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got := stored.ArchivedAt(); got == nil || !got.Equal(archivedAt) {
		t.Fatalf("ArchivedAt() = %v, want first archive time %v", got, archivedAt)
	}
}

func TestPostgresServiceRepositoryReturnsNotFound(t *testing.T) {
	repository := setupRepository(t)

	_, err := repository.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, application.ErrServiceNotFound) {
		t.Fatalf("GetByID() error = %v, want ErrServiceNotFound", err)
	}
}

func setupRepository(t *testing.T) *ServiceRepository {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("pool.Ping() error = %v", err)
	}
	if err := migration.New(pool, migrations.Files).Up(ctx); err != nil {
		t.Fatalf("migration Up() error = %v", err)
	}

	transaction, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin() error = %v", err)
	}
	t.Cleanup(func() {
		if err := transaction.Rollback(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Errorf("transaction.Rollback() error = %v", err)
		}
	})
	return NewServiceRepository(transaction)
}

func newTestService(t *testing.T, serviceKey string) domain.Service {
	t.Helper()
	service, err := domain.NewService(domain.Registration{
		ServiceKey:    serviceKey,
		DisplayName:   "Catalog test service",
		Description:   "Created by repository integration tests",
		OwnerTeam:     "platform-team",
		RepositoryURL: "https://github.com/example/" + serviceKey,
		Lifecycle:     domain.LifecycleDevelopment,
	}, time.Now().UTC().Add(-time.Minute))
	if err != nil {
		t.Fatalf("domain.NewService() error = %v", err)
	}
	return service
}

func containsService(services []domain.Service, id uuid.UUID) bool {
	for _, service := range services {
		if service.ID() == id {
			return true
		}
	}
	return false
}
