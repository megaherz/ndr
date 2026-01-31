package postgres

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

// MigrationRunner handles database migrations using golang-migrate
type MigrationRunner struct {
	db     *DB
	logger *logrus.Logger
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *DB, logger *logrus.Logger) *MigrationRunner {
	return &MigrationRunner{
		db:     db,
		logger: logger,
	}
}

// RunMigrations executes all pending migrations using golang-migrate
func (m *MigrationRunner) RunMigrations(ctx context.Context, migrationsDir string) error {
	m.logger.Info("Starting database migrations")

	// Create a database driver instance
	driver, err := postgres.WithInstance(m.db.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path for migrations
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Create migrate instance
	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := migrator.Close(); sourceErr != nil || dbErr != nil {
			m.logger.WithFields(logrus.Fields{
				"source_error": sourceErr,
				"db_error":     dbErr,
			}).Warn("Error closing migrator")
		}
	}()

	// Run migrations
	if err := migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			m.logger.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	m.logger.Info("Database migrations completed successfully")
	return nil
}

// GetAppliedMigrations returns a list of applied migrations for status checking
func (m *MigrationRunner) GetAppliedMigrations(ctx context.Context) ([]string, error) {

	// Query the schema_migrations table directly
	// The golang-migrate package uses 'schema_migrations' table by default
	query := `SELECT version FROM schema_migrations ORDER BY version`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		// If table doesn't exist, return empty list
		if err.Error() == `pq: relation "schema_migrations" does not exist` {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer func() {
		_ = rows.Close() // Ignore close error in defer
	}()

	var migrations []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		migrations = append(migrations, version)
	}

	return migrations, rows.Err()
}

// GetMigrationVersion returns the current migration version
func (m *MigrationRunner) GetMigrationVersion(migrationsDir string) (uint, bool, error) {
	// Create a database driver instance
	driver, err := postgres.WithInstance(m.db.DB.DB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path for migrations
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Create migrate instance
	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := migrator.Close(); sourceErr != nil || dbErr != nil {
			m.logger.WithFields(logrus.Fields{
				"source_error": sourceErr,
				"db_error":     dbErr,
			}).Warn("Error closing migrator")
		}
	}()

	version, dirty, err := migrator.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return 0, false, nil // No migrations applied yet
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
