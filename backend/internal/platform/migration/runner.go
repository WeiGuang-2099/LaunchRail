package migration

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const advisoryLockKey int64 = 4_962_821_019

var migrationFilename = regexp.MustCompile(`^(\d{6})_([a-z0-9_]+)\.(up|down)\.sql$`)

type Runner struct {
	pool  *pgxpool.Pool
	files fs.FS
}

type migrationFile struct {
	version  int64
	name     string
	upPath   string
	downPath string
}

func New(pool *pgxpool.Pool, files fs.FS) *Runner {
	return &Runner{pool: pool, files: files}
}

func (r *Runner) Up(ctx context.Context) error {
	migrations, err := loadMigrationFiles(r.files)
	if err != nil {
		return err
	}

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()

	if err := acquireLock(ctx, connection); err != nil {
		return err
	}
	defer releaseLock(connection)

	if err := ensureMigrationsTable(ctx, connection); err != nil {
		return err
	}

	applied, err := appliedMigrations(ctx, connection)
	if err != nil {
		return err
	}
	if err := validateAppliedMigrations(migrations, applied); err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, exists := applied[migration.version]; exists {
			continue
		}
		if err := r.applyUp(ctx, connection, migration); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) Down(ctx context.Context, steps int) error {
	if steps <= 0 {
		return fmt.Errorf("migration steps must be greater than zero")
	}

	migrations, err := loadMigrationFiles(r.files)
	if err != nil {
		return err
	}
	migrationsByVersion := make(map[int64]migrationFile, len(migrations))
	for _, migration := range migrations {
		migrationsByVersion[migration.version] = migration
	}

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()

	if err := acquireLock(ctx, connection); err != nil {
		return err
	}
	defer releaseLock(connection)

	if err := ensureMigrationsTable(ctx, connection); err != nil {
		return err
	}

	rows, err := connection.Query(ctx, `
		SELECT version, name
		FROM schema_migrations
		ORDER BY version DESC
		LIMIT $1`, steps)
	if err != nil {
		return fmt.Errorf("list migrations to roll back: %w", err)
	}
	type appliedMigration struct {
		version int64
		name    string
	}
	applied := make([]appliedMigration, 0, steps)
	for rows.Next() {
		var item appliedMigration
		if err := rows.Scan(&item.version, &item.name); err != nil {
			rows.Close()
			return fmt.Errorf("scan migration to roll back: %w", err)
		}
		applied = append(applied, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate migrations to roll back: %w", err)
	}
	rows.Close()

	for _, item := range applied {
		migration, exists := migrationsByVersion[item.version]
		if !exists || migration.name != item.name {
			return fmt.Errorf("applied migration %06d_%s is not available", item.version, item.name)
		}
		if err := r.applyDown(ctx, connection, migration); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) applyUp(ctx context.Context, connection *pgxpool.Conn, migration migrationFile) error {
	contents, err := fs.ReadFile(r.files, migration.upPath)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", migration.upPath, err)
	}

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %06d: %w", migration.version, err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Conn().PgConn().Exec(ctx, string(contents)).ReadAll(); err != nil {
		return fmt.Errorf("apply migration %06d_%s: %w", migration.version, migration.name, err)
	}
	if _, err := transaction.Exec(ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		migration.version,
		migration.name,
	); err != nil {
		return fmt.Errorf("record migration %06d_%s: %w", migration.version, migration.name, err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %06d_%s: %w", migration.version, migration.name, err)
	}
	return nil
}

func (r *Runner) applyDown(ctx context.Context, connection *pgxpool.Conn, migration migrationFile) error {
	contents, err := fs.ReadFile(r.files, migration.downPath)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", migration.downPath, err)
	}

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration rollback %06d: %w", migration.version, err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Conn().PgConn().Exec(ctx, string(contents)).ReadAll(); err != nil {
		return fmt.Errorf("roll back migration %06d_%s: %w", migration.version, migration.name, err)
	}
	if _, err := transaction.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, migration.version); err != nil {
		return fmt.Errorf("remove migration record %06d_%s: %w", migration.version, migration.name, err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration rollback %06d_%s: %w", migration.version, migration.name, err)
	}
	return nil
}

func loadMigrationFiles(files fs.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("list migration files: %w", err)
	}

	migrations := make(map[int64]migrationFile)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFilename.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse migration version in %s: %w", entry.Name(), err)
		}

		migration := migrations[version]
		if migration.name != "" && migration.name != matches[2] {
			return nil, fmt.Errorf("migration version %06d has conflicting names", version)
		}
		migration.version = version
		migration.name = matches[2]
		if matches[3] == "up" {
			migration.upPath = entry.Name()
		} else {
			migration.downPath = entry.Name()
		}
		migrations[version] = migration
	}

	result := make([]migrationFile, 0, len(migrations))
	for _, migration := range migrations {
		if migration.upPath == "" || migration.downPath == "" {
			return nil, fmt.Errorf("migration %06d_%s must have up and down files", migration.version, migration.name)
		}
		result = append(result, migration)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].version < result[j].version })
	return result, nil
}

func acquireLock(ctx context.Context, connection *pgxpool.Conn) error {
	if _, err := connection.Exec(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	return nil
}

func releaseLock(connection *pgxpool.Conn) {
	_, _ = connection.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, advisoryLockKey)
}

func ensureMigrationsTable(ctx context.Context, connection *pgxpool.Conn) error {
	_, err := connection.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func appliedMigrations(ctx context.Context, connection *pgxpool.Conn) (map[int64]string, error) {
	rows, err := connection.Query(ctx, `SELECT version, name FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]string)
	for rows.Next() {
		var version int64
		var name string
		if err := rows.Scan(&version, &name); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = name
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return applied, nil
}

func validateAppliedMigrations(migrations []migrationFile, applied map[int64]string) error {
	available := make(map[int64]string, len(migrations))
	for _, migration := range migrations {
		available[migration.version] = migration.name
	}
	for version, name := range applied {
		availableName, exists := available[version]
		if !exists || availableName != name {
			return fmt.Errorf("applied migration %06d_%s is not available", version, name)
		}
	}
	return nil
}

var _ pgx.Tx = (pgx.Tx)(nil)
