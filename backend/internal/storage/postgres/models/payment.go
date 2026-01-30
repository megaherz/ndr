package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Payment represents TON deposit and withdrawal intents
type Payment struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	UserID          uuid.UUID       `db:"user_id" json:"user_id"`
	PaymentType     PaymentType     `db:"payment_type" json:"payment_type"`
	Status          PaymentStatus   `db:"status" json:"status"`
	TonAmount       decimal.Decimal `db:"ton_amount" json:"ton_amount"`
	FuelAmount      decimal.Decimal `db:"fuel_amount" json:"fuel_amount"`
	TonTxHash       *string         `db:"ton_tx_hash" json:"ton_tx_hash,omitempty"`
	ClientRequestID uuid.UUID       `db:"client_request_id" json:"client_request_id"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

// PaymentType represents the direction of the payment (PostgreSQL ENUM: payment_type)
type PaymentType string

const (
	PaymentTypeDeposit    PaymentType = "DEPOSIT"
	PaymentTypeWithdrawal PaymentType = "WITHDRAWAL"
)

// String returns the string representation
func (pt PaymentType) String() string {
	return string(pt)
}

// IsValid checks if the payment type is valid
func (pt PaymentType) IsValid() bool {
	switch pt {
	case PaymentTypeDeposit, PaymentTypeWithdrawal:
		return true
	}
	return false
}

// PaymentStatus represents the current status of a payment (PostgreSQL ENUM: payment_status_type)
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusConfirmed PaymentStatus = "CONFIRMED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

// String returns the string representation
func (ps PaymentStatus) String() string {
	return string(ps)
}

// IsValid checks if the payment status is valid
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusConfirmed, PaymentStatusFailed:
		return true
	}
	return false
}

// IsDeposit returns true if this is a deposit payment
func (p *Payment) IsDeposit() bool {
	return p.PaymentType == PaymentTypeDeposit
}

// IsWithdrawal returns true if this is a withdrawal payment
func (p *Payment) IsWithdrawal() bool {
	return p.PaymentType == PaymentTypeWithdrawal
}

// IsPending returns true if the payment is still pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsConfirmed returns true if the payment is confirmed
func (p *Payment) IsConfirmed() bool {
	return p.Status == PaymentStatusConfirmed
}

// IsFailed returns true if the payment has failed
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}
