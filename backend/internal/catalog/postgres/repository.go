package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/application"
	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const serviceColumns = `
	id,
	service_key,
	display_name,
	description,
	owner_team,
	repository_url,
	lifecycle,
	created_at,
	updated_at,
	archived_at`

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type ServiceRepository struct {
	database DBTX
}

func NewServiceRepository(database DBTX) *ServiceRepository {
	return &ServiceRepository{database: database}
}

func (r *ServiceRepository) Create(ctx context.Context, service domain.Service) error {
	_, err := r.database.Exec(ctx, `
		INSERT INTO services (
			id, service_key, display_name, description, owner_team,
			repository_url, lifecycle, created_at, updated_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		service.ID(),
		service.ServiceKey(),
		service.DisplayName(),
		service.Description(),
		service.OwnerTeam(),
		service.RepositoryURL(),
		service.Lifecycle(),
		service.CreatedAt(),
		service.UpdatedAt(),
		service.ArchivedAt(),
	)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" && postgresError.ConstraintName == "services_service_key_unique" {
			return application.ErrServiceKeyConflict
		}
		return fmt.Errorf("insert service: %w", err)
	}
	return nil
}

func (r *ServiceRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Service, error) {
	row := r.database.QueryRow(ctx, `SELECT `+serviceColumns+` FROM services WHERE id = $1`, id)
	service, err := scanService(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Service{}, application.ErrServiceNotFound
	}
	if err != nil {
		return domain.Service{}, fmt.Errorf("get service: %w", err)
	}
	return service, nil
}

func (r *ServiceRepository) List(ctx context.Context, options application.ListOptions) (application.ServicePage, error) {
	if !options.Status.Valid() {
		return application.ServicePage{}, fmt.Errorf("list services: invalid status %q", options.Status)
	}
	if options.Limit <= 0 || options.Limit > 100 {
		return application.ServicePage{}, fmt.Errorf("list services: limit must be between 1 and 100")
	}
	if options.Offset < 0 {
		return application.ServicePage{}, fmt.Errorf("list services: offset must not be negative")
	}

	whereClause := listWhereClause(options.Status)
	var total int64
	if err := r.database.QueryRow(ctx, `SELECT count(*) FROM services `+whereClause).Scan(&total); err != nil {
		return application.ServicePage{}, fmt.Errorf("count services: %w", err)
	}

	rows, err := r.database.Query(ctx, `
		SELECT `+serviceColumns+`
		FROM services `+whereClause+`
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`, options.Limit, options.Offset)
	if err != nil {
		return application.ServicePage{}, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	services := make([]domain.Service, 0, options.Limit)
	for rows.Next() {
		service, err := scanService(rows)
		if err != nil {
			return application.ServicePage{}, fmt.Errorf("scan listed service: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return application.ServicePage{}, fmt.Errorf("iterate services: %w", err)
	}

	return application.ServicePage{Services: services, Total: total}, nil
}

func (r *ServiceRepository) Update(ctx context.Context, service domain.Service) error {
	if service.Archived() {
		return domain.ErrArchivedService
	}

	result, err := r.database.Exec(ctx, `
		UPDATE services
		SET display_name = $2,
			description = $3,
			owner_team = $4,
			repository_url = $5,
			lifecycle = $6,
			updated_at = $7
		WHERE id = $1 AND archived_at IS NULL`,
		service.ID(),
		service.DisplayName(),
		service.Description(),
		service.OwnerTeam(),
		service.RepositoryURL(),
		service.Lifecycle(),
		service.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	if result.RowsAffected() == 1 {
		return nil
	}
	return r.classifyUnavailableService(ctx, service.ID())
}

func (r *ServiceRepository) Archive(ctx context.Context, id uuid.UUID, archivedAt time.Time) error {
	if archivedAt.IsZero() {
		return fmt.Errorf("archive service: archivedAt must not be zero")
	}

	result, err := r.database.Exec(ctx, `
		UPDATE services
		SET archived_at = $2, updated_at = $2
		WHERE id = $1 AND archived_at IS NULL`, id, archivedAt.UTC())
	if err != nil {
		return fmt.Errorf("archive service: %w", err)
	}
	if result.RowsAffected() == 1 {
		return nil
	}

	var exists bool
	if err := r.database.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM services WHERE id = $1)`, id).Scan(&exists); err != nil {
		return fmt.Errorf("check archived service: %w", err)
	}
	if !exists {
		return application.ErrServiceNotFound
	}
	return nil
}

func (r *ServiceRepository) classifyUnavailableService(ctx context.Context, id uuid.UUID) error {
	var archived bool
	err := r.database.QueryRow(ctx, `SELECT archived_at IS NOT NULL FROM services WHERE id = $1`, id).Scan(&archived)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ErrServiceNotFound
	}
	if err != nil {
		return fmt.Errorf("classify service update: %w", err)
	}
	if archived {
		return domain.ErrArchivedService
	}
	return fmt.Errorf("update service: no row was changed")
}

func listWhereClause(status application.ListStatus) string {
	switch status {
	case application.ListStatusActive:
		return "WHERE archived_at IS NULL"
	case application.ListStatusArchived:
		return "WHERE archived_at IS NOT NULL"
	default:
		return ""
	}
}

type rowScanner interface {
	Scan(...any) error
}

func scanService(row rowScanner) (domain.Service, error) {
	var record domain.ServiceRecord
	var id string
	var lifecycle string
	if err := row.Scan(
		&id,
		&record.ServiceKey,
		&record.DisplayName,
		&record.Description,
		&record.OwnerTeam,
		&record.RepositoryURL,
		&lifecycle,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.ArchivedAt,
	); err != nil {
		return domain.Service{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.Service{}, fmt.Errorf("parse service ID: %w", err)
	}
	record.ID = parsedID
	record.Lifecycle = domain.Lifecycle(lifecycle)

	service, err := domain.RestoreService(record)
	if err != nil {
		return domain.Service{}, fmt.Errorf("restore service: %w", err)
	}
	return service, nil
}

var _ application.ServiceRepository = (*ServiceRepository)(nil)
