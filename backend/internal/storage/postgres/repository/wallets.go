package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"backend/internal/storage/postgres/models"
)

// WalletRepository defines the interface for wallet data access
type WalletRepository interface {
	// GetByUserID retrieves a wallet by user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error)
	
	// Create creates a new wallet for a user
	Create(ctx context.Context, wallet *models.Wallet) error
	
	// UpdateBalances updates wallet balances atomically
	UpdateBalances(ctx context.Context, userID uuid.UUID, tonDelta, fuelDelta, burnDelta decimal.Decimal) error
	
	// IncrementRookieRaces increments the rookie races completed counter
	IncrementRookieRaces(ctx context.Context, userID uuid.UUID) error
	
	// SetTONWalletAddress sets the connected TON wallet address
	SetTONWalletAddress(ctx context.Context, userID uuid.UUID, address string) error
}

// walletRepository implements WalletRepository
type walletRepository struct {
	db *sqlx.DB
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *sqlx.DB) WalletRepository {
	return &walletRepository{db: db}
}

// GetByUserID retrieves a wallet by user ID
func (r *walletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `
		SELECT user_id, ton_balance, fuel_balance, burn_balance, 
		       rookie_races_completed, ton_wallet_address, created_at, updated_at
		FROM wallets 
		WHERE user_id = $1`
	
	err := r.db.GetContext(ctx, wallet, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return wallet, nil
}

// Create creates a new wallet for a user
func (r *walletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (user_id, ton_balance, fuel_balance, burn_balance, 
		                    rookie_races_completed, ton_wallet_address, created_at, updated_at)
		VALUES (:user_id, :ton_balance, :fuel_balance, :burn_balance, 
		        :rookie_races_completed, :ton_wallet_address, :created_at, :updated_at)`
	
	_, err := r.db.NamedExecContext(ctx, query, wallet)
	return err
}

// UpdateBalances updates wallet balances atomically
func (r *walletRepository) UpdateBalances(ctx context.Context, userID uuid.UUID, tonDelta, fuelDelta, burnDelta decimal.Decimal) error {
	query := `
		UPDATE wallets 
		SET ton_balance = ton_balance + $2,
		    fuel_balance = fuel_balance + $3,
		    burn_balance = burn_balance + $4,
		    updated_at = NOW()
		WHERE user_id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID, tonDelta, fuelDelta, burnDelta)
	return err
}

// IncrementRookieRaces increments the rookie races completed counter
func (r *walletRepository) IncrementRookieRaces(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE wallets 
		SET rookie_races_completed = rookie_races_completed + 1,
		    updated_at = NOW()
		WHERE user_id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// SetTONWalletAddress sets the connected TON wallet address
func (r *walletRepository) SetTONWalletAddress(ctx context.Context, userID uuid.UUID, address string) error {
	query := `
		UPDATE wallets 
		SET ton_wallet_address = $2,
		    updated_at = NOW()
		WHERE user_id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID, address)
	return err
}