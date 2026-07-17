package application

import (
	"context"
	"fmt"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/domain"
	"github.com/google/uuid"
)

type Clock func() time.Time

type Catalog struct {
	repository ServiceRepository
	now        Clock
}

func NewCatalog(repository ServiceRepository, now Clock) *Catalog {
	if now == nil {
		now = time.Now
	}
	return &Catalog{repository: repository, now: now}
}

func (c *Catalog) Register(ctx context.Context, registration domain.Registration) (domain.Service, error) {
	service, err := domain.NewService(registration, c.now())
	if err != nil {
		return domain.Service{}, err
	}
	if err := c.repository.Create(ctx, service); err != nil {
		return domain.Service{}, fmt.Errorf("register service: %w", err)
	}
	return service, nil
}

func (c *Catalog) Get(ctx context.Context, id uuid.UUID) (domain.Service, error) {
	service, err := c.repository.GetByID(ctx, id)
	if err != nil {
		return domain.Service{}, fmt.Errorf("get service: %w", err)
	}
	return service, nil
}

func (c *Catalog) List(ctx context.Context, options ListOptions) (ServicePage, error) {
	page, err := c.repository.List(ctx, options)
	if err != nil {
		return ServicePage{}, fmt.Errorf("list services: %w", err)
	}
	return page, nil
}

func (c *Catalog) Update(ctx context.Context, id uuid.UUID, metadata domain.MutableMetadata) (domain.Service, error) {
	service, err := c.repository.GetByID(ctx, id)
	if err != nil {
		return domain.Service{}, fmt.Errorf("load service for update: %w", err)
	}
	if err := service.Update(metadata, c.now()); err != nil {
		return domain.Service{}, err
	}
	if err := c.repository.Update(ctx, service); err != nil {
		return domain.Service{}, fmt.Errorf("persist service update: %w", err)
	}
	return service, nil
}

func (c *Catalog) Archive(ctx context.Context, id uuid.UUID) error {
	service, err := c.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("load service for archive: %w", err)
	}
	if err := service.Archive(c.now()); err != nil {
		return err
	}
	archivedAt := service.ArchivedAt()
	if archivedAt == nil {
		return fmt.Errorf("archive service: archive timestamp is missing")
	}
	if err := c.repository.Archive(ctx, id, *archivedAt); err != nil {
		return fmt.Errorf("persist service archive: %w", err)
	}
	return nil
}
