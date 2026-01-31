package account

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"backend/internal/constants"
	"backend/internal/storage/postgres/models"
	"backend/internal/storage/postgres/repository"
)

// LedgerOperations handles ledger entry operations
type LedgerOperations interface {
	// DebitFuel debits FUEL from a user's account
	DebitFuel(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error
	
	// CreditFuel credits FUEL to a user's account
	CreditFuel(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error
	
	// CreditBurn credits BURN to a user's account
	CreditBurn(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error
	
	// DebitSystemWallet debits FUEL from a system wallet
	DebitSystemWallet(ctx context.Context, walletName string, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error
	
	// CreditSystemWallet credits FUEL to a system wallet
	CreditSystemWallet(ctx context.Context, walletName string, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error
	
	// RecordEntry records a generic ledger entry
	RecordEntry(ctx context.Context, entry *models.LedgerEntry) error
	
	// RecordMatchEntries records multiple ledger entries for a match atomically
	RecordMatchEntries(ctx context.Context, entries []*models.LedgerEntry) error
	
	// TransferFuel transfers FUEL between users
	TransferFuel(ctx context.Context, fromUserID, toUserID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error
}

// ledgerOperations implements LedgerOperations
type ledgerOperations struct {
	ledgerRepo repository.LedgerRepository
	walletRepo repository.WalletRepository
	logger     *logrus.Logger
}

// NewLedgerOperations creates a new ledger operations handler
func NewLedgerOperations(
	ledgerRepo repository.LedgerRepository,
	walletRepo repository.WalletRepository,
	logger *logrus.Logger,
) LedgerOperations {
	return &ledgerOperations{
		ledgerRepo: ledgerRepo,
		walletRepo: walletRepo,
		logger:     logger,
	}
}

// DebitFuel debits FUEL from a user's account
func (l *ledgerOperations) DebitFuel(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("debit amount must be positive")
	}
	
	// Create debit entry (negative amount)
	entry := &models.LedgerEntry{
		UserID:        &userID,
		SystemWallet:  sql.NullString{},
		Currency:      constants.CurrencyFUEL,
		Amount:        amount.Neg(), // Negative for debit
		OperationType: operationType,
		ReferenceID:   referenceID,
		Description:   sql.NullString{String: description, Valid: description != ""},
		CreatedAt:     time.Now(),
	}
	
	// Record entry and update wallet balance
	err := l.recordEntryAndUpdateBalance(ctx, entry)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"user_id":        userID,
			"amount":         amount,
			"operation_type": operationType,
			"error":          err,
		}).Error("Failed to debit FUEL")
		return fmt.Errorf("failed to debit FUEL: %w", err)
	}
	
	return nil
}

// CreditFuel credits FUEL to a user's account
func (l *ledgerOperations) CreditFuel(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("credit amount must be positive")
	}
	
	// Create credit entry (positive amount)
	entry := &models.LedgerEntry{
		UserID:        &userID,
		SystemWallet:  sql.NullString{},
		Currency:      constants.CurrencyFUEL,
		Amount:        amount, // Positive for credit
		OperationType: operationType,
		ReferenceID:   referenceID,
		Description:   sql.NullString{String: description, Valid: description != ""},
		CreatedAt:     time.Now(),
	}
	
	// Record entry and update wallet balance
	err := l.recordEntryAndUpdateBalance(ctx, entry)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"user_id":        userID,
			"amount":         amount,
			"operation_type": operationType,
			"error":          err,
		}).Error("Failed to credit FUEL")
		return fmt.Errorf("failed to credit FUEL: %w", err)
	}
	
	return nil
}

// CreditBurn credits BURN to a user's account
func (l *ledgerOperations) CreditBurn(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("credit amount must be positive")
	}
	
	// Create credit entry (positive amount)
	entry := &models.LedgerEntry{
		UserID:        &userID,
		SystemWallet:  sql.NullString{},
		Currency:      constants.CurrencyBURN,
		Amount:        amount, // Positive for credit
		OperationType: operationType,
		ReferenceID:   referenceID,
		Description:   sql.NullString{String: description, Valid: description != ""},
		CreatedAt:     time.Now(),
	}
	
	// Record entry and update wallet balance
	err := l.recordEntryAndUpdateBalance(ctx, entry)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"user_id":        userID,
			"amount":         amount,
			"operation_type": operationType,
			"error":          err,
		}).Error("Failed to credit BURN")
		return fmt.Errorf("failed to credit BURN: %w", err)
	}
	
	return nil
}

// DebitSystemWallet debits FUEL from a system wallet
func (l *ledgerOperations) DebitSystemWallet(ctx context.Context, walletName string, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("debit amount must be positive")
	}
	
	// Create debit entry (negative amount)
	entry := &models.LedgerEntry{
		UserID:        nil,
		SystemWallet:  sql.NullString{String: walletName, Valid: true},
		Currency:      constants.CurrencyFUEL,
		Amount:        amount.Neg(), // Negative for debit
		OperationType: operationType,
		ReferenceID:   referenceID,
		Description:   sql.NullString{String: description, Valid: description != ""},
		CreatedAt:     time.Now(),
	}
	
	// Record entry (system wallets don't have direct balance updates)
	err := l.ledgerRepo.CreateEntry(ctx, entry)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"wallet_name":    walletName,
			"amount":         amount,
			"operation_type": operationType,
			"error":          err,
		}).Error("Failed to debit system wallet")
		return fmt.Errorf("failed to debit system wallet: %w", err)
	}
	
	return nil
}

// CreditSystemWallet credits FUEL to a system wallet
func (l *ledgerOperations) CreditSystemWallet(ctx context.Context, walletName string, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("credit amount must be positive")
	}
	
	// Create credit entry (positive amount)
	entry := &models.LedgerEntry{
		UserID:        nil,
		SystemWallet:  sql.NullString{String: walletName, Valid: true},
		Currency:      constants.CurrencyFUEL,
		Amount:        amount, // Positive for credit
		OperationType: operationType,
		ReferenceID:   referenceID,
		Description:   sql.NullString{String: description, Valid: description != ""},
		CreatedAt:     time.Now(),
	}
	
	// Record entry (system wallets don't have direct balance updates)
	err := l.ledgerRepo.CreateEntry(ctx, entry)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"wallet_name":    walletName,
			"amount":         amount,
			"operation_type": operationType,
			"error":          err,
		}).Error("Failed to credit system wallet")
		return fmt.Errorf("failed to credit system wallet: %w", err)
	}
	
	return nil
}

// RecordEntry records a generic ledger entry
func (l *ledgerOperations) RecordEntry(ctx context.Context, entry *models.LedgerEntry) error {
	// Validate entry
	if entry.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if entry.OperationType == "" {
		return fmt.Errorf("operation type is required")
	}
	if entry.Amount.IsZero() {
		return fmt.Errorf("amount cannot be zero")
	}
	
	// Set created timestamp if not set
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	
	// Record entry and update balance if it's a user entry
	if entry.UserID != nil {
		return l.recordEntryAndUpdateBalance(ctx, entry)
	}
	
	// System wallet entry - just record
	return l.ledgerRepo.CreateEntry(ctx, entry)
}

// RecordMatchEntries records multiple ledger entries for a match atomically
func (l *ledgerOperations) RecordMatchEntries(ctx context.Context, entries []*models.LedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}
	
	// Validate all entries have the same reference ID (match ID)
	var matchID *uuid.UUID
	for i, entry := range entries {
		if i == 0 {
			matchID = entry.ReferenceID
		} else if entry.ReferenceID == nil || (matchID != nil && *entry.ReferenceID != *matchID) {
			return fmt.Errorf("all match entries must have the same reference ID")
		}
	}
	
	// Record all entries atomically
	err := l.ledgerRepo.CreateEntries(ctx, entries)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"match_id":     matchID,
			"entry_count":  len(entries),
			"error":        err,
		}).Error("Failed to record match entries")
		return fmt.Errorf("failed to record match entries: %w", err)
	}
	
	// Update wallet balances for user entries
	for _, entry := range entries {
		if entry.UserID != nil {
			err := l.updateWalletBalance(ctx, *entry.UserID, entry.Currency, entry.Amount)
			if err != nil {
				l.logger.WithFields(logrus.Fields{
					"user_id":  *entry.UserID,
					"currency": entry.Currency,
					"amount":   entry.Amount,
					"error":    err,
				}).Error("Failed to update wallet balance after match entry")
				// Continue with other entries, but log the error
			}
		}
	}
	
	return nil
}

// TransferFuel transfers FUEL between users
func (l *ledgerOperations) TransferFuel(ctx context.Context, fromUserID, toUserID uuid.UUID, amount decimal.Decimal, operationType string, referenceID *uuid.UUID, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("transfer amount must be positive")
	}
	
	// Create debit entry for sender
	debitEntry := &models.LedgerEntry{
		UserID:        &fromUserID,
		SystemWallet:  sql.NullString{},
		Currency:      constants.CurrencyFUEL,
		Amount:        amount.Neg(), // Negative for debit
		OperationType: operationType,
		ReferenceID:   referenceID,
		Description:   sql.NullString{String: fmt.Sprintf("Transfer to %s: %s", toUserID, description), Valid: true},
		CreatedAt:     time.Now(),
	}
	
	// Create credit entry for receiver
	creditEntry := &models.LedgerEntry{
		UserID:        &toUserID,
		SystemWallet:  sql.NullString{},
		Currency:      constants.CurrencyFUEL,
		Amount:        amount, // Positive for credit
		OperationType: operationType,
		ReferenceID:   referenceID,
		Description:   sql.NullString{String: fmt.Sprintf("Transfer from %s: %s", fromUserID, description), Valid: true},
		CreatedAt:     time.Now(),
	}
	
	// Record both entries atomically
	entries := []*models.LedgerEntry{debitEntry, creditEntry}
	err := l.ledgerRepo.CreateEntries(ctx, entries)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"from_user_id":   fromUserID,
			"to_user_id":     toUserID,
			"amount":         amount,
			"operation_type": operationType,
			"error":          err,
		}).Error("Failed to transfer FUEL")
		return fmt.Errorf("failed to transfer FUEL: %w", err)
	}
	
	// Update wallet balances
	err = l.updateWalletBalance(ctx, fromUserID, "FUEL", amount.Neg())
	if err != nil {
		return fmt.Errorf("failed to update sender balance: %w", err)
	}
	
	err = l.updateWalletBalance(ctx, toUserID, "FUEL", amount)
	if err != nil {
		return fmt.Errorf("failed to update receiver balance: %w", err)
	}
	
	return nil
}

// recordEntryAndUpdateBalance records a ledger entry and updates the wallet balance
func (l *ledgerOperations) recordEntryAndUpdateBalance(ctx context.Context, entry *models.LedgerEntry) error {
	// Record the ledger entry
	err := l.ledgerRepo.CreateEntry(ctx, entry)
	if err != nil {
		return err
	}
	
	// Update wallet balance if it's a user entry
	if entry.UserID != nil {
		err = l.updateWalletBalance(ctx, *entry.UserID, entry.Currency, entry.Amount)
		if err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}
	}
	
	return nil
}

// updateWalletBalance updates the wallet balance for a specific currency
func (l *ledgerOperations) updateWalletBalance(ctx context.Context, userID uuid.UUID, currency string, delta decimal.Decimal) error {
	var tonDelta, fuelDelta, burnDelta decimal.Decimal
	
	switch currency {
	case constants.CurrencyTON:
		tonDelta = delta
	case constants.CurrencyFUEL:
		fuelDelta = delta
	case constants.CurrencyBURN:
		burnDelta = delta
	default:
		return fmt.Errorf("unsupported currency: %s", currency)
	}
	
	return l.walletRepo.UpdateBalances(ctx, userID, tonDelta, fuelDelta, burnDelta)
}