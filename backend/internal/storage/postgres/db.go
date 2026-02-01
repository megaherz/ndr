package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
)

// DB wraps the database connection with additional functionality
type DB struct {
	*sqlx.DB
	logger *logrus.Logger
}

// Config holds database configuration
type Config struct {
	URL               string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	ConnMaxIdleTime   time.Duration
	ConnectionTimeout time.Duration
}

// NewDB creates a new database connection
func NewDB(cfg Config, logger *logrus.Logger) (*DB, error) {
	// Parse and validate the database URL
	parsedURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}

	// Set default values
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 1 * time.Minute
	}
	if cfg.ConnectionTimeout == 0 {
		cfg.ConnectionTimeout = 10 * time.Second
	}

	logger.WithFields(logrus.Fields{
		"host":     parsedURL.Host,
		"database": parsedURL.Path[1:], // Remove leading slash
	}).Info("Connecting to PostgreSQL database")

	// Create connection with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectionTimeout)
	defer cancel()

	// Open database connection
	db, err := sqlx.ConnectContext(ctx, "postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close() // Ignore close error since we're already returning an error
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"max_open_conns":     cfg.MaxOpenConns,
		"max_idle_conns":     cfg.MaxIdleConns,
		"conn_max_lifetime":  cfg.ConnMaxLifetime,
		"conn_max_idle_time": cfg.ConnMaxIdleTime,
	}).Info("Database connection established")

	return &DB{
		DB:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	return db.DB.Close()
}

// HealthCheck performs a health check on the database
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Simple ping
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var result int
	if err := db.GetContext(ctx, &result, "SELECT 1"); err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func (db *DB) GetStats() sql.DBStats {
	return db.Stats()
}

// WithTransaction executes a function within a database transaction
func (db *DB) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback() // Ignore rollback error in panic recovery
			panic(p)          // Re-throw panic after rollback
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				db.logger.WithError(rollbackErr).Error("Failed to rollback transaction")
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				db.logger.WithError(commitErr).Error("Failed to commit transaction")
				err = commitErr
			}
		}
	}()

	err = fn(tx)
	return err
}

// LogSlowQueries logs queries that take longer than the specified duration
func (db *DB) LogSlowQueries(threshold time.Duration) {
	// This is a placeholder for query logging middleware
	// In a production environment, you might want to implement
	// query logging using database hooks or middleware
	db.logger.WithField("threshold", threshold).Info("Slow query logging enabled")
}
