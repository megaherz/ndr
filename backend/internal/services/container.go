package services

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/auth"
	"github.com/megaherz/ndr/internal/centrifugo"
	"github.com/megaherz/ndr/internal/config"
	"github.com/megaherz/ndr/internal/modules/account"
	authservice "github.com/megaherz/ndr/internal/modules/auth"
	"github.com/megaherz/ndr/internal/modules/gameengine"
	"github.com/megaherz/ndr/internal/modules/gateway"
	"github.com/megaherz/ndr/internal/modules/matchmaker"
	"github.com/megaherz/ndr/internal/storage/postgres"
	"github.com/megaherz/ndr/internal/storage/postgres/repository"
	"github.com/megaherz/ndr/internal/storage/redis"
)

// Container holds all application services and dependencies
type Container struct {
	// Configuration
	Config *config.Config

	// Storage
	DB          *postgres.DB
	RedisClient *redis.Client

	// Repositories
	UserRepo             repository.UserRepository
	WalletRepo           repository.WalletRepository
	LedgerRepo           repository.LedgerRepository
	MatchRepo            repository.MatchRepository
	MatchParticipantRepo repository.MatchParticipantRepository
	MatchSettlementRepo  repository.MatchSettlementRepository

	// Utilities
	JWTManager       *auth.JWTManager
	CentrifugoClient *centrifugo.Client

	// Services
	AuthService       authservice.AuthService
	AccountService    account.AccountService
	GameEngineService gameengine.GameEngineService
	MatchmakerService matchmaker.MatchmakerService

	// Logger
	Logger *logrus.Logger
}

// NewContainer creates and initializes a new service container
func NewContainer(cfg *config.Config, logger *logrus.Logger) (*Container, error) {
	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	// Initialize in dependency order
	if err := container.initializeStorage(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	if err := container.initializeRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := container.initializeUtilities(); err != nil {
		return nil, fmt.Errorf("failed to initialize utilities: %w", err)
	}

	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	logger.Info("Service container initialized successfully")
	return container, nil
}

// initializeStorage sets up database and Redis connections
func (c *Container) initializeStorage() error {
	// Initialize PostgreSQL
	dbConfig := postgres.Config{
		URL:               c.Config.DatabaseURL,
		MaxOpenConns:      25,
		MaxIdleConns:      5,
		ConnMaxLifetime:   5 * time.Minute,
		ConnMaxIdleTime:   1 * time.Minute,
		ConnectionTimeout: 10 * time.Second,
	}

	db, err := postgres.NewDB(dbConfig, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	c.DB = db

	// Run database migrations
	if err := c.runMigrations(); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Initialize Redis
	redisConfig, err := parseRedisURL(c.Config.RedisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	redisClient, err := redis.NewClient(*redisConfig, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	c.RedisClient = redisClient

	return nil
}

// initializeRepositories creates all repository instances
func (c *Container) initializeRepositories() error {
	c.UserRepo = repository.NewUserRepository(c.DB.DB)
	c.WalletRepo = repository.NewWalletRepository(c.DB.DB)
	c.LedgerRepo = repository.NewLedgerRepository(c.DB.DB)
	c.MatchRepo = repository.NewMatchRepository(c.DB.DB)
	c.MatchParticipantRepo = repository.NewMatchParticipantRepository(c.DB.DB)
	c.MatchSettlementRepo = repository.NewMatchSettlementRepository(c.DB.DB)

	c.Logger.Info("Repositories initialized")
	return nil
}

// initializeUtilities creates utility instances
func (c *Container) initializeUtilities() error {
	// Initialize JWT Manager
	c.JWTManager = auth.NewJWTManager(c.Config.JWTSecret, "ndr-api")

	// Initialize Centrifugo Client
	centrifugoClient, err := centrifugo.NewClient(centrifugo.Config{
		GRPCAddr: c.Config.CentrifugoGRPCAddr,
		APIKey:   c.Config.CentrifugoAPIKey,
	}, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize Centrifugo client: %w", err)
	}
	c.CentrifugoClient = centrifugoClient

	c.Logger.Info("Utilities initialized")
	return nil
}

// initializeServices creates all service instances
func (c *Container) initializeServices() error {
	// Auth Service - needs user repo, wallet repo, JWT manager
	c.AuthService = authservice.NewAuthService(
		c.UserRepo,
		c.WalletRepo,
		c.JWTManager,
		"", // Bot token - should be added to config if needed
		c.Logger,
	)

	// Account Service - needs wallet repo, ledger repo
	c.AccountService = account.NewAccountService(
		c.WalletRepo,
		c.LedgerRepo,
		c.Logger,
	)

	// Game Engine Service - needs match repos and participant repo
	c.GameEngineService = gameengine.NewGameEngineService(
		c.MatchRepo,
		c.MatchParticipantRepo,
		c.Logger,
	)

	// Matchmaker Service - needs queue operations, account service, and publisher
	queueOps := matchmaker.NewQueueOperations(c.RedisClient.GetClient())
	publisher := gateway.NewCentrifugoPublisher(c.CentrifugoClient, c.Logger)
	c.MatchmakerService = matchmaker.NewMatchmakerService(
		queueOps,
		c.AccountService,
		publisher,
		c.Logger,
	)

	c.Logger.Info("Services initialized")
	return nil
}

// Close gracefully shuts down all connections and services
func (c *Container) Close() error {
	var errors []error

	// Close Centrifugo client
	if c.CentrifugoClient != nil {
		if err := c.CentrifugoClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Centrifugo client: %w", err))
		}
	}

	// Close Redis connection
	if c.RedisClient != nil {
		if err := c.RedisClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Redis client: %w", err))
		}
	}

	// Close database connection
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if len(errors) > 0 {
		c.Logger.WithField("errors", errors).Error("Errors occurred during container shutdown")
		return fmt.Errorf("container shutdown errors: %v", errors)
	}

	c.Logger.Info("Service container closed successfully")
	return nil
}

// HealthCheck performs health checks on all critical services
func (c *Container) HealthCheck(ctx context.Context) error {
	// Check database
	if err := c.DB.HealthCheck(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check Redis
	if err := c.RedisClient.GetClient().Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	// Check Centrifugo (if needed)
	// This would depend on the Centrifugo client implementation

	return nil
}

// parseRedisURL parses a Redis URL into a Redis config
func parseRedisURL(redisURL string) (*redis.Config, error) {
	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	config := &redis.Config{
		Addr: u.Host,
		DB:   0, // Default database
	}

	// Extract password if present
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			config.Password = password
		}
	}

	// Extract database number from path
	if u.Path != "" && u.Path != "/" {
		// Remove leading slash and parse as integer
		dbStr := u.Path[1:]
		if db, err := strconv.Atoi(dbStr); err == nil {
			config.DB = db
		}
	}

	return config, nil
}

// runMigrations executes database migrations
func (c *Container) runMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	migrationRunner := postgres.NewMigrationRunner(c.DB, c.Logger)

	// Determine migrations directory path
	// This assumes the migrations are in the standard location relative to the binary
	migrationsDir := "internal/storage/postgres/migrations"

	return migrationRunner.RunMigrations(ctx, migrationsDir)
}
