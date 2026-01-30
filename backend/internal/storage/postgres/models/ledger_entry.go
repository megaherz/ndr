package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// LedgerEntry represents an immutable audit trail of currency movements
type LedgerEntry struct {
	ID            int64           `db:"id" json:"id"`
	UserID        *uuid.UUID      `db:"user_id" json:"user_id,omitempty"`
	SystemWallet  *string         `db:"system_wallet" json:"system_wallet,omitempty"`
	Currency      Currency        `db:"currency" json:"currency"`
	Amount        decimal.Decimal `db:"amount" json:"amount"`
	OperationType OperationType   `db:"operation_type" json:"operation_type"`
	ReferenceID   *uuid.UUID      `db:"reference_id" json:"reference_id,omitempty"`
	Description   *string         `db:"description" json:"description,omitempty"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
}

// Currency types
type Currency string

const (
	CurrencyTON  Currency = "TON"
	CurrencyFUEL Currency = "FUEL"
	CurrencyBURN Currency = "BURN"
)

// Operation types for ledger entries
type OperationType string

const (
	OperationDeposit          OperationType = "DEPOSIT"
	OperationWithdrawal       OperationType = "WITHDRAWAL"
	OperationMatchBuyin       OperationType = "MATCH_BUYIN"
	OperationMatchPrize       OperationType = "MATCH_PRIZE"
	OperationMatchRake        OperationType = "MATCH_RAKE"
	OperationMatchBurnReward  OperationType = "MATCH_BURN_REWARD"
	OperationInitialBalance   OperationType = "INITIAL_BALANCE"
)