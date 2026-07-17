package application

import (
	"context"
	"errors"
	"time"

	"github.com/WeiGuang-2099/LaunchRail/backend/internal/catalog/domain"
	"github.com/google/uuid"
)

var (
	ErrServiceNotFound    = errors.New("service not found")
	ErrServiceKeyConflict = errors.New("service key already exists")
)

type ListStatus string

const (
	ListStatusActive   ListStatus = "active"
	ListStatusArchived ListStatus = "archived"
	ListStatusAll      ListStatus = "all"
)

type ListOptions struct {
	Status ListStatus
	Limit  int
	Offset int
}

type ServicePage struct {
	Services []domain.Service
	Total    int64
}

type ServiceRepository interface {
	Create(context.Context, domain.Service) error
	GetByID(context.Context, uuid.UUID) (domain.Service, error)
	List(context.Context, ListOptions) (ServicePage, error)
	Update(context.Context, domain.Service) error
	Archive(context.Context, uuid.UUID, time.Time) error
}

func (s ListStatus) Valid() bool {
	switch s {
	case ListStatusActive, ListStatusArchived, ListStatusAll:
		return true
	default:
		return false
	}
}
