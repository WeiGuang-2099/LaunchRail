INSERT INTO services (
    id,
    service_key,
    display_name,
    description,
    owner_team,
    repository_url,
    lifecycle,
    created_at,
    updated_at,
    archived_at
)
VALUES (
    '00000000-0000-4000-8000-000000000001',
    'launchrail-e2e',
    'LaunchRail E2E Service',
    'Deterministic service record for full-stack smoke tests.',
    'platform-team',
    'https://github.com/WeiGuang-2099/LaunchRail',
    'production',
    '2026-07-20T00:00:00Z',
    '2026-07-20T00:00:00Z',
    NULL
)
ON CONFLICT (service_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    owner_team = EXCLUDED.owner_team,
    repository_url = EXCLUDED.repository_url,
    lifecycle = EXCLUDED.lifecycle,
    updated_at = EXCLUDED.updated_at,
    archived_at = NULL;
