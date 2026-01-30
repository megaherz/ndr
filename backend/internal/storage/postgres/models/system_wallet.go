package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// SystemWallet represents special wallets for house and rake pools
type SystemWallet struct {
	WalletName  string          `db:"wallet_name" json:"wallet_name"`
	FuelBalance decimal.Decimal `db:"fuel_balance" json:"fuel_balance"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

// System wallet names
const (
	HouseFuelWallet = "HOUSE_FUEL"
	RakeFuelWallet  = "RAKE_FUEL"
)
