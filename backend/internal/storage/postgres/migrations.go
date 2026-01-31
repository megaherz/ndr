package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// MigrationRunner handles database migrations
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

// RunMigrations executes all pending migrations
func (m *MigrationRunner) RunMigrations(ctx context.Context, migrationsDir string) error {
	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := m.getMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, file := range migrationFiles {
		if strings.HasSuffix(file, ".down.sql") {
			continue // Skip down migrations
		}

		migrationName := m.getMigrationName(file)
		if appliedMigrations[migrationName] {
			continue // Already applied
		}

		if err := m.applyMigration(ctx, migrationsDir, file, migrationName); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migrationName, err)
		}
	}

	return nil
}

// ensureMigrationsTable creates the migrations tracking table if it doesn't exist
func (m *MigrationRunner) ensureMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getMigrationFiles returns a sorted list of migration files
func (m *MigrationRunner) getMigrationFiles(migrationsDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, filepath.Base(path))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files to ensure consistent order
	sort.Strings(files)
	return files, nil
}

// getAppliedMigrations returns a map of applied migration names
func (m *MigrationRunner) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := `SELECT version FROM schema_migrations`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// getMigrationName extracts the migration name from a filename
func (m *MigrationRunner) getMigrationName(filename string) string {
	// Remove .up.sql suffix
	name := strings.TrimSuffix(filename, ".up.sql")
	return name
}

// applyMigration applies a single migration
func (m *MigrationRunner) applyMigration(ctx context.Context, migrationsDir, filename, migrationName string) error {
	m.logger.WithField("migration", migrationName).Info("Applying migration")

	// Read migration file
	migrationPath := filepath.Join(migrationsDir, filename)
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration in transaction
	return m.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Execute migration SQL
		_, err := tx.ExecContext(ctx, string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}

		// Record migration as applied
		_, err = tx.ExecContext(ctx, 
			`INSERT INTO schema_migrations (version) VALUES ($1)`, 
			migrationName)
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		m.logger.WithField("migration", migrationName).Info("Migration applied successfully")
		return nil
	})
}

// GetAppliedMigrations returns a list of applied migrations for status checking
func (m *MigrationRunner) GetAppliedMigrations(ctx context.Context) ([]string, error) {
	query := `SELECT version FROM schema_migrations ORDER BY applied_at`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations = append(migrations, version)
	}

	return migrations, rows.Err()
}