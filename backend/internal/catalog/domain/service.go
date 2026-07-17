package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Lifecycle string

const (
	LifecycleDevelopment Lifecycle = "development"
	LifecycleProduction  Lifecycle = "production"
	LifecycleDeprecated  Lifecycle = "deprecated"
)

type Registration struct {
	ServiceKey    string
	DisplayName   string
	Description   string
	OwnerTeam     string
	RepositoryURL string
	Lifecycle     Lifecycle
}

type MutableMetadata struct {
	DisplayName   string
	Description   string
	OwnerTeam     string
	RepositoryURL string
	Lifecycle     Lifecycle
}

type ServiceRecord struct {
	ID            uuid.UUID
	ServiceKey    string
	DisplayName   string
	Description   string
	OwnerTeam     string
	RepositoryURL string
	Lifecycle     Lifecycle
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ArchivedAt    *time.Time
}

type Service struct {
	id            uuid.UUID
	serviceKey    string
	displayName   string
	description   string
	ownerTeam     string
	repositoryURL string
	lifecycle     Lifecycle
	createdAt     time.Time
	updatedAt     time.Time
	archivedAt    *time.Time
}

func NewService(registration Registration, now time.Time) (Service, error) {
	registration = normalizeRegistration(registration)
	validationErrors := validateRegistration(registration)
	if now.IsZero() {
		validationErrors.add("createdAt", "must not be zero")
	}
	if err := validationErrors.orNil(); err != nil {
		return Service{}, err
	}

	now = normalizeTime(now)
	return Service{
		id:            uuid.New(),
		serviceKey:    registration.ServiceKey,
		displayName:   registration.DisplayName,
		description:   registration.Description,
		ownerTeam:     registration.OwnerTeam,
		repositoryURL: registration.RepositoryURL,
		lifecycle:     registration.Lifecycle,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func RestoreService(record ServiceRecord) (Service, error) {
	registration := normalizeRegistration(Registration{
		ServiceKey:    record.ServiceKey,
		DisplayName:   record.DisplayName,
		Description:   record.Description,
		OwnerTeam:     record.OwnerTeam,
		RepositoryURL: record.RepositoryURL,
		Lifecycle:     record.Lifecycle,
	})
	validationErrors := validateRegistration(registration)
	if record.ID == uuid.Nil {
		validationErrors.add("id", "must be a valid UUID")
	}
	if record.CreatedAt.IsZero() {
		validationErrors.add("createdAt", "must not be zero")
	}
	if record.UpdatedAt.IsZero() || record.UpdatedAt.Before(record.CreatedAt) {
		validationErrors.add("updatedAt", "must not be before createdAt")
	}
	if record.ArchivedAt != nil && record.ArchivedAt.Before(record.CreatedAt) {
		validationErrors.add("archivedAt", "must not be before createdAt")
	}
	if err := validationErrors.orNil(); err != nil {
		return Service{}, err
	}

	var archivedAt *time.Time
	if record.ArchivedAt != nil {
		value := normalizeTime(*record.ArchivedAt)
		archivedAt = &value
	}

	return Service{
		id:            record.ID,
		serviceKey:    registration.ServiceKey,
		displayName:   registration.DisplayName,
		description:   registration.Description,
		ownerTeam:     registration.OwnerTeam,
		repositoryURL: registration.RepositoryURL,
		lifecycle:     registration.Lifecycle,
		createdAt:     normalizeTime(record.CreatedAt),
		updatedAt:     normalizeTime(record.UpdatedAt),
		archivedAt:    archivedAt,
	}, nil
}

func (s *Service) Update(metadata MutableMetadata, now time.Time) error {
	if s.Archived() {
		return ErrArchivedService
	}

	metadata = normalizeMetadata(metadata)
	validationErrors := validateMetadata(metadata)
	if now.IsZero() || now.Before(s.updatedAt) {
		validationErrors.add("updatedAt", "must not be before the current updatedAt")
	}
	if err := validationErrors.orNil(); err != nil {
		return err
	}

	s.displayName = metadata.DisplayName
	s.description = metadata.Description
	s.ownerTeam = metadata.OwnerTeam
	s.repositoryURL = metadata.RepositoryURL
	s.lifecycle = metadata.Lifecycle
	s.updatedAt = normalizeTime(now)
	return nil
}

func (s *Service) Archive(now time.Time) error {
	if s.Archived() {
		return nil
	}
	if now.IsZero() || now.Before(s.updatedAt) {
		return &ValidationErrors{Violations: []FieldViolation{{
			Field:   "archivedAt",
			Message: "must not be before the current updatedAt",
		}}}
	}

	now = normalizeTime(now)
	s.archivedAt = &now
	s.updatedAt = now
	return nil
}

func (s Service) ID() uuid.UUID         { return s.id }
func (s Service) ServiceKey() string    { return s.serviceKey }
func (s Service) DisplayName() string   { return s.displayName }
func (s Service) Description() string   { return s.description }
func (s Service) OwnerTeam() string     { return s.ownerTeam }
func (s Service) RepositoryURL() string { return s.repositoryURL }
func (s Service) Lifecycle() Lifecycle  { return s.lifecycle }
func (s Service) CreatedAt() time.Time  { return s.createdAt }
func (s Service) UpdatedAt() time.Time  { return s.updatedAt }
func (s Service) Archived() bool        { return s.archivedAt != nil }
func (s Service) ArchivedAt() *time.Time {
	if s.archivedAt == nil {
		return nil
	}
	value := *s.archivedAt
	return &value
}

func normalizeRegistration(registration Registration) Registration {
	registration.ServiceKey = strings.TrimSpace(registration.ServiceKey)
	metadata := normalizeMetadata(MutableMetadata{
		DisplayName:   registration.DisplayName,
		Description:   registration.Description,
		OwnerTeam:     registration.OwnerTeam,
		RepositoryURL: registration.RepositoryURL,
		Lifecycle:     registration.Lifecycle,
	})
	registration.DisplayName = metadata.DisplayName
	registration.Description = metadata.Description
	registration.OwnerTeam = metadata.OwnerTeam
	registration.RepositoryURL = metadata.RepositoryURL
	registration.Lifecycle = metadata.Lifecycle
	return registration
}

func normalizeMetadata(metadata MutableMetadata) MutableMetadata {
	metadata.DisplayName = strings.TrimSpace(metadata.DisplayName)
	metadata.Description = strings.TrimSpace(metadata.Description)
	metadata.OwnerTeam = strings.TrimSpace(metadata.OwnerTeam)
	metadata.RepositoryURL = strings.TrimSpace(metadata.RepositoryURL)
	metadata.Lifecycle = Lifecycle(strings.TrimSpace(string(metadata.Lifecycle)))
	return metadata
}

func validateRegistration(registration Registration) *ValidationErrors {
	validationErrors := validateMetadata(MutableMetadata{
		DisplayName:   registration.DisplayName,
		Description:   registration.Description,
		OwnerTeam:     registration.OwnerTeam,
		RepositoryURL: registration.RepositoryURL,
		Lifecycle:     registration.Lifecycle,
	})

	if !validServiceKey(registration.ServiceKey) {
		validationErrors.add("serviceKey", "must match ^[a-z][a-z0-9-]{0,62}$")
	}
	return validationErrors
}

func validateMetadata(metadata MutableMetadata) *ValidationErrors {
	validationErrors := &ValidationErrors{}
	validateRequiredLength(validationErrors, "displayName", metadata.DisplayName, 120)
	if utf8.RuneCountInString(metadata.Description) > 2000 {
		validationErrors.add("description", "must be at most 2000 characters")
	}
	validateRequiredLength(validationErrors, "ownerTeam", metadata.OwnerTeam, 120)
	validateRequiredLength(validationErrors, "repositoryUrl", metadata.RepositoryURL, 512)
	if metadata.RepositoryURL != "" {
		if err := validateGitHubRepositoryURL(metadata.RepositoryURL); err != nil {
			validationErrors.add("repositoryUrl", err.Error())
		}
	}
	if !metadata.Lifecycle.Valid() {
		validationErrors.add("lifecycle", "must be one of development, production, or deprecated")
	}
	return validationErrors
}

func validateRequiredLength(validationErrors *ValidationErrors, field, value string, maximum int) {
	length := utf8.RuneCountInString(value)
	if length == 0 {
		validationErrors.add(field, "is required")
		return
	}
	if length > maximum {
		validationErrors.add(field, fmt.Sprintf("must be at most %d characters", maximum))
	}
}

func validServiceKey(value string) bool {
	if len(value) == 0 || len(value) > 63 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			character == '-' {
			continue
		}
		return false
	}
	return true
}

func validateGitHubRepositoryURL(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return fmt.Errorf("must be a valid HTTPS URL")
	}
	if parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "github.com") || parsed.Port() != "" {
		return fmt.Errorf("must use https://github.com")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("must not contain credentials, query parameters, or fragments")
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || strings.TrimSuffix(parts[1], ".git") == "" {
		return fmt.Errorf("must identify a GitHub owner and repository")
	}
	return nil
}

func (l Lifecycle) Valid() bool {
	switch l {
	case LifecycleDevelopment, LifecycleProduction, LifecycleDeprecated:
		return true
	default:
		return false
	}
}

func normalizeTime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}
