package repository

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

// TestDBHelper provides shared database setup and teardown for integration tests
type TestDBHelper struct {
	Pool     *dockertest.Pool
	Resource *dockertest.Resource
	DB       *sqlx.DB
	t        *testing.T
}

// NewTestDBHelper creates a new test database helper
func NewTestDBHelper(t *testing.T) *TestDBHelper {
	return &TestDBHelper{t: t}
}

// SetupDatabase starts a PostgreSQL container and applies migrations
func (h *TestDBHelper) SetupDatabase() {
	var err error

	// Create dockertest pool
	h.Pool, err = dockertest.NewPool("")
	require.NoError(h.t, err)

	// Set shorter retry timeout for faster tests
	h.Pool.MaxWait = 120 * time.Second

	// Start PostgreSQL container with version 17
	h.Resource, err = h.Pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=testpass",
			"POSTGRES_USER=testuser",
			"POSTGRES_DB=testdb",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(h.t, err)

	// Get the container's host and port
	hostAndPort := h.Resource.GetHostPort("5432/tcp")
	databaseURL := fmt.Sprintf("postgres://testuser:testpass@%s/testdb?sslmode=disable", hostAndPort)

	log.Printf("Connecting to database on %s", hostAndPort)

	// Wait for the database to be ready
	err = h.Pool.Retry(func() error {
		h.DB, err = sqlx.Connect("postgres", databaseURL)
		if err != nil {
			return err
		}
		return h.DB.Ping()
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	require.NoError(h.t, err)

	// Apply migrations
	err = h.applyMigrations()
	require.NoError(h.t, err)
}

// TeardownDatabase closes the database connection and removes the container
func (h *TestDBHelper) TeardownDatabase() {
	if h.DB != nil {
		_ = h.DB.Close() // Ignore close error during cleanup
	}
	if h.Resource != nil {
		err := h.Pool.Purge(h.Resource)
		require.NoError(h.t, err)
	}
}

// CleanupTables truncates all tables for a clean test state
func (h *TestDBHelper) CleanupTables(tables ...string) {
	for _, table := range tables {
		_, err := h.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(h.t, err)
	}
}

// applyMigrations reads and applies all migration files in the correct order
func (h *TestDBHelper) applyMigrations() error {
	migrationsPath := "../migrations"

	// Read migration files
	var migrationFiles []string
	err := filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files to ensure correct order
	sort.Strings(migrationFiles)

	// Apply each migration
	for _, migrationFile := range migrationFiles {
		log.Printf("Applying migration: %s", migrationFile)

		content, err := os.ReadFile(migrationFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
		}

		// Execute the migration
		_, err = h.DB.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationFile, err)
		}
	}

	log.Printf("Applied %d migrations successfully", len(migrationFiles))
	return nil
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
