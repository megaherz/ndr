package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Wallet represents a player's currency balances
type Wallet struct {
	UserID               uuid.UUID       `db:"user_id" json:"user_id"`
	TonBalance           decimal.Decimal `db:"ton_balance" json:"ton_balance"`
	FuelBalance          decimal.Decimal `db:"fuel_balance" json:"fuel_balance"`
	BurnBalance          decimal.Decimal `db:"burn_balance" json:"burn_balance"`
	RookieRacesCompleted int             `db:"rookie_races_completed" json:"rookie_races_completed"`
	TonWalletAddress     *string         `db:"ton_wallet_address" json:"ton_wallet_address,omitempty"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at" json:"updated_at"`
}
