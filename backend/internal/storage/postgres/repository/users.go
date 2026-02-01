package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/megaherz/ndr/internal/storage/postgres/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)

	// GetByTelegramID retrieves a user by Telegram ID
	GetByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)

	// UpdateTelegramInfo updates user's Telegram information
	UpdateTelegramInfo(ctx context.Context, userID uuid.UUID, username, firstName, lastName, photoURL string) error

	// GetOrCreateByTelegramID gets an existing user or creates a new one
	GetOrCreateByTelegramID(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) (*models.User, error)

	// List retrieves users with pagination
	List(ctx context.Context, limit, offset int) ([]*models.User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, telegram_id, telegram_username, telegram_first_name, 
		                  telegram_last_name, telegram_photo_url, created_at, updated_at)
		VALUES (:id, :telegram_id, :telegram_username, :telegram_first_name, 
		        :telegram_last_name, :telegram_photo_url, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, telegram_id, telegram_username, telegram_first_name, 
		       telegram_last_name, telegram_photo_url, created_at, updated_at
		FROM users 
		WHERE id = $1`

	err := r.db.GetContext(ctx, user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// GetByTelegramID retrieves a user by Telegram ID
func (r *userRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, telegram_id, telegram_username, telegram_first_name, 
		       telegram_last_name, telegram_photo_url, created_at, updated_at
		FROM users 
		WHERE telegram_id = $1`

	err := r.db.GetContext(ctx, user, query, telegramID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// UpdateTelegramInfo updates user's Telegram information
func (r *userRepository) UpdateTelegramInfo(ctx context.Context, userID uuid.UUID, username, firstName, lastName, photoURL string) error {
	query := `
		UPDATE users 
		SET telegram_username = $2, 
		    telegram_first_name = $3, 
		    telegram_last_name = $4,
		    telegram_photo_url = $5,
		    updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID,
		sql.NullString{String: username, Valid: username != ""},
		firstName,
		sql.NullString{String: lastName, Valid: lastName != ""},
		sql.NullString{String: photoURL, Valid: photoURL != ""})
	return err
}

// GetOrCreateByTelegramID gets an existing user or creates a new one
func (r *userRepository) GetOrCreateByTelegramID(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) (*models.User, error) {
	// First, try to get existing user
	existingUser, err := r.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		// Update Telegram info in case it changed
		err = r.UpdateTelegramInfo(ctx, existingUser.ID, username, firstName, lastName, photoURL)
		if err != nil {
			return nil, err
		}

		// Return updated user data
		return r.GetByID(ctx, existingUser.ID)
	}

	// User doesn't exist, create new one
	var usernamePtr, lastNamePtr, photoURLPtr *string
	if username != "" {
		usernamePtr = &username
	}
	if lastName != "" {
		lastNamePtr = &lastName
	}
	if photoURL != "" {
		photoURLPtr = &photoURL
	}

	newUser := &models.User{
		ID:                uuid.New(),
		TelegramID:        telegramID,
		TelegramUsername:  usernamePtr,
		TelegramFirstName: firstName,
		TelegramLastName:  lastNamePtr,
		TelegramPhotoURL:  photoURLPtr,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = r.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// List retrieves users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users := []*models.User{}
	query := `
		SELECT id, telegram_id, telegram_username, telegram_first_name, 
		       telegram_last_name, telegram_photo_url, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	return users, err
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`

	err := r.db.GetContext(ctx, &count, query)
	return count, err
}
