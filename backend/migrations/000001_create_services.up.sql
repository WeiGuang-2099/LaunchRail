CREATE TABLE services (
    id UUID PRIMARY KEY,
    service_key VARCHAR(63) NOT NULL,
    display_name VARCHAR(120) NOT NULL,
    description VARCHAR(2000) NOT NULL DEFAULT '',
    owner_team VARCHAR(120) NOT NULL,
    repository_url VARCHAR(512) NOT NULL,
    lifecycle VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    archived_at TIMESTAMPTZ,
    CONSTRAINT services_service_key_unique UNIQUE (service_key),
    CONSTRAINT services_service_key_format CHECK (service_key ~ '^[a-z][a-z0-9-]{0,62}$'),
    CONSTRAINT services_display_name_not_blank CHECK (btrim(display_name) <> ''),
    CONSTRAINT services_description_length CHECK (char_length(description) <= 2000),
    CONSTRAINT services_owner_team_not_blank CHECK (btrim(owner_team) <> ''),
    CONSTRAINT services_repository_url_not_blank CHECK (btrim(repository_url) <> ''),
    CONSTRAINT services_lifecycle_valid CHECK (lifecycle IN ('development', 'production', 'deprecated')),
    CONSTRAINT services_updated_after_creation CHECK (updated_at >= created_at),
    CONSTRAINT services_archived_after_creation CHECK (archived_at IS NULL OR archived_at >= created_at)
);

CREATE INDEX idx_services_active_created_at
    ON services (created_at DESC, id DESC)
    WHERE archived_at IS NULL;

CREATE INDEX idx_services_archived_at
    ON services (archived_at DESC, id DESC)
    WHERE archived_at IS NOT NULL;
