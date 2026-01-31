package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"github.com/megaherz/ndr/internal/constants"
	"github.com/megaherz/ndr/internal/storage/postgres/models"
)

// LedgerRepository defines the interface for ledger entry data access
type LedgerRepository interface {
	// CreateEntry creates a new ledger entry
	CreateEntry(ctx context.Context, entry *models.LedgerEntry) error
	
	// CreateEntries creates multiple ledger entries in a transaction
	CreateEntries(ctx context.Context, entries []*models.LedgerEntry) error
	
	// GetUserEntries retrieves ledger entries for a user with pagination
	GetUserEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.LedgerEntry, error)
	
	// GetMatchEntries retrieves all ledger entries for a match
	GetMatchEntries(ctx context.Context, matchID uuid.UUID) ([]*models.LedgerEntry, error)
	
	// GetUserBalance calculates current balance for a user and currency
	GetUserBalance(ctx context.Context, userID uuid.UUID, currency string) (decimal.Decimal, error)
	
	// GetSystemWalletBalance calculates current balance for a system wallet
	GetSystemWalletBalance(ctx context.Context, walletName string) (decimal.Decimal, error)
	
	// ValidateMatchLedgerBalance validates that all entries for a match sum to zero
	ValidateMatchLedgerBalance(ctx context.Context, matchID uuid.UUID) (bool, error)
}

// ledgerRepository implements LedgerRepository
type ledgerRepository struct {
	db *sqlx.DB
}

// NewLedgerRepository creates a new ledger repository
func NewLedgerRepository(db *sqlx.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

// CreateEntry creates a new ledger entry
func (r *ledgerRepository) CreateEntry(ctx context.Context, entry *models.LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries (user_id, system_wallet, currency, amount, 
		                           operation_type, reference_id, description, created_at)
		VALUES (:user_id, :system_wallet, :currency, :amount, 
		        :operation_type, :reference_id, :description, :created_at)`
	
	_, err := r.db.NamedExecContext(ctx, query, entry)
	return err
}

// CreateEntries creates multiple ledger entries in a transaction
func (r *ledgerRepository) CreateEntries(ctx context.Context, entries []*models.LedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	query := `
		INSERT INTO ledger_entries (user_id, system_wallet, currency, amount, 
		                           operation_type, reference_id, description, created_at)
		VALUES (:user_id, :system_wallet, :currency, :amount, 
		        :operation_type, :reference_id, :description, :created_at)`
	
	for _, entry := range entries {
		_, err := tx.NamedExecContext(ctx, query, entry)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

// GetUserEntries retrieves ledger entries for a user with pagination
func (r *ledgerRepository) GetUserEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.LedgerEntry, error) {
	entries := []*models.LedgerEntry{}
	query := `
		SELECT id, user_id, system_wallet, currency, amount, operation_type, 
		       reference_id, description, created_at
		FROM ledger_entries 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	err := r.db.SelectContext(ctx, &entries, query, userID, limit, offset)
	return entries, err
}

// GetMatchEntries retrieves all ledger entries for a match
func (r *ledgerRepository) GetMatchEntries(ctx context.Context, matchID uuid.UUID) ([]*models.LedgerEntry, error) {
	entries := []*models.LedgerEntry{}
	query := `
		SELECT id, user_id, system_wallet, currency, amount, operation_type, 
		       reference_id, description, created_at
		FROM ledger_entries 
		WHERE reference_id = $1
		ORDER BY created_at ASC`
	
	err := r.db.SelectContext(ctx, &entries, query, matchID)
	return entries, err
}

// GetUserBalance calculates current balance for a user and currency
func (r *ledgerRepository) GetUserBalance(ctx context.Context, userID uuid.UUID, currency string) (decimal.Decimal, error) {
	var balance sql.NullString
	query := `
		SELECT COALESCE(SUM(amount), 0)::text
		FROM ledger_entries 
		WHERE user_id = $1 AND currency = $2`
	
	err := r.db.GetContext(ctx, &balance, query, userID, currency)
	if err != nil {
		return decimal.Zero, err
	}
	
	if !balance.Valid {
		return decimal.Zero, nil
	}
	
	return decimal.NewFromString(balance.String)
}

// GetSystemWalletBalance calculates current balance for a system wallet
func (r *ledgerRepository) GetSystemWalletBalance(ctx context.Context, walletName string) (decimal.Decimal, error) {
	var balance sql.NullString
	query := `
		SELECT COALESCE(SUM(amount), 0)::text
		FROM ledger_entries 
		WHERE system_wallet = $1 AND currency = $2`
	
	err := r.db.GetContext(ctx, &balance, query, walletName, constants.CurrencyFUEL)
	if err != nil {
		return decimal.Zero, err
	}
	
	if !balance.Valid {
		return decimal.Zero, nil
	}
	
	return decimal.NewFromString(balance.String)
}

// ValidateMatchLedgerBalance validates that all entries for a match sum to zero
func (r *ledgerRepository) ValidateMatchLedgerBalance(ctx context.Context, matchID uuid.UUID) (bool, error) {
	var balance sql.NullString
	query := `
		SELECT COALESCE(SUM(amount), 0)::text
		FROM ledger_entries 
		WHERE reference_id = $1`
	
	err := r.db.GetContext(ctx, &balance, query, matchID)
	if err != nil {
		return false, err
	}
	
	if !balance.Valid {
		return true, nil // No entries means balanced
	}
	
	balanceDecimal, err := decimal.NewFromString(balance.String)
	if err != nil {
		return false, err
	}
	
	return balanceDecimal.IsZero(), nil
}