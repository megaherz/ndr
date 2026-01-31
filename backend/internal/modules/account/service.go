package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"backend/internal/constants"
	"backend/internal/storage/postgres/models"
	"backend/internal/storage/postgres/repository"
)

// AccountService handles account and wallet operations
type AccountService interface {
	// GetWallet retrieves wallet information for a user
	GetWallet(ctx context.Context, userID uuid.UUID) (*WalletInfo, error)
	
	// GetBalance retrieves current balance for a user and currency
	GetBalance(ctx context.Context, userID uuid.UUID, currency string) (decimal.Decimal, error)
	
	// HasSufficientBalance checks if user has enough balance for an operation
	HasSufficientBalance(ctx context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (bool, error)
	
	// GetSystemWalletBalance retrieves balance for a system wallet
	GetSystemWalletBalance(ctx context.Context, walletName string) (decimal.Decimal, error)
	
	// GetTransactionHistory retrieves transaction history for a user
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.LedgerEntry, error)
}

// WalletInfo represents comprehensive wallet information
type WalletInfo struct {
	UserID               uuid.UUID       `json:"user_id"`
	TONBalance           decimal.Decimal `json:"ton_balance"`
	FuelBalance          decimal.Decimal `json:"fuel_balance"`
	BurnBalance          decimal.Decimal `json:"burn_balance"`
	RookieRacesCompleted int             `json:"rookie_races_completed"`
	TONWalletAddress     *string         `json:"ton_wallet_address,omitempty"`
	LeagueAccess         LeagueAccess    `json:"league_access"`
}

// LeagueAccess represents which leagues the user can access
type LeagueAccess struct {
	Rookie   LeagueStatus `json:"rookie"`
	Street   LeagueStatus `json:"street"`
	Pro      LeagueStatus `json:"pro"`
	TopFuel  LeagueStatus `json:"top_fuel"`
}

// LeagueStatus represents the status of a league for a user
type LeagueStatus struct {
	Accessible bool            `json:"accessible"`
	BuyinCost  decimal.Decimal `json:"buyin_cost"`
	Reason     string          `json:"reason,omitempty"` // Why not accessible
}

// Use league constants from constants package
var (
	RookieBuyin  = constants.LeagueBuyins[constants.LeagueRookie]
	StreetBuyin  = constants.LeagueBuyins[constants.LeagueStreet]
	ProBuyin     = constants.LeagueBuyins[constants.LeaguePro]
	TopFuelBuyin = constants.LeagueBuyins[constants.LeagueTopFuel]
)

// accountService implements AccountService
type accountService struct {
	walletRepo repository.WalletRepository
	ledgerRepo repository.LedgerRepository
	logger     *logrus.Logger
}

// NewAccountService creates a new account service
func NewAccountService(
	walletRepo repository.WalletRepository,
	ledgerRepo repository.LedgerRepository,
	logger *logrus.Logger,
) AccountService {
	return &accountService{
		walletRepo: walletRepo,
		ledgerRepo: ledgerRepo,
		logger:     logger,
	}
}

// GetWallet retrieves wallet information for a user
func (s *accountService) GetWallet(ctx context.Context, userID uuid.UUID) (*WalletInfo, error) {
	// Get wallet from database
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get wallet")
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found for user %s", userID)
	}
	
	// Calculate league access
	leagueAccess := s.calculateLeagueAccess(wallet)
	
	// Build wallet info
	walletInfo := &WalletInfo{
		UserID:               wallet.UserID,
		TONBalance:           wallet.TONBalance,
		FuelBalance:          wallet.FuelBalance,
		BurnBalance:          wallet.BurnBalance,
		RookieRacesCompleted: wallet.RookieRacesCompleted,
		LeagueAccess:         leagueAccess,
	}
	
	// Add TON wallet address if present
	if wallet.TONWalletAddress.Valid {
		walletInfo.TONWalletAddress = &wallet.TONWalletAddress.String
	}
	
	return walletInfo, nil
}

// GetBalance retrieves current balance for a user and currency
func (s *accountService) GetBalance(ctx context.Context, userID uuid.UUID, currency string) (decimal.Decimal, error) {
	balance, err := s.ledgerRepo.GetUserBalance(ctx, userID, currency)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id":  userID,
			"currency": currency,
			"error":    err,
		}).Error("Failed to get user balance")
		return decimal.Zero, fmt.Errorf("failed to get balance: %w", err)
	}
	
	return balance, nil
}

// HasSufficientBalance checks if user has enough balance for an operation
func (s *accountService) HasSufficientBalance(ctx context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (bool, error) {
	balance, err := s.GetBalance(ctx, userID, currency)
	if err != nil {
		return false, err
	}
	
	return balance.GreaterThanOrEqual(amount), nil
}

// GetSystemWalletBalance retrieves balance for a system wallet
func (s *accountService) GetSystemWalletBalance(ctx context.Context, walletName string) (decimal.Decimal, error) {
	balance, err := s.ledgerRepo.GetSystemWalletBalance(ctx, walletName)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"wallet_name": walletName,
			"error":       err,
		}).Error("Failed to get system wallet balance")
		return decimal.Zero, fmt.Errorf("failed to get system wallet balance: %w", err)
	}
	
	return balance, nil
}

// GetTransactionHistory retrieves transaction history for a user
func (s *accountService) GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.LedgerEntry, error) {
	entries, err := s.ledgerRepo.GetUserEntries(ctx, userID, limit, offset)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"limit":   limit,
			"offset":  offset,
			"error":   err,
		}).Error("Failed to get transaction history")
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	
	return entries, nil
}

// calculateLeagueAccess determines which leagues a user can access
func (s *accountService) calculateLeagueAccess(wallet *models.Wallet) LeagueAccess {
	access := LeagueAccess{}
	
	// Rookie league
	if wallet.RookieRacesCompleted >= 3 {
		access.Rookie = LeagueStatus{
			Accessible: false,
			BuyinCost:  RookieBuyin,
			Reason:     "Maximum 3 rookie races completed",
		}
	} else if wallet.FuelBalance.LessThan(RookieBuyin) {
		access.Rookie = LeagueStatus{
			Accessible: false,
			BuyinCost:  RookieBuyin,
			Reason:     "Insufficient FUEL balance",
		}
	} else {
		access.Rookie = LeagueStatus{
			Accessible: true,
			BuyinCost:  RookieBuyin,
		}
	}
	
	// Street league
	if wallet.FuelBalance.LessThan(StreetBuyin) {
		access.Street = LeagueStatus{
			Accessible: false,
			BuyinCost:  StreetBuyin,
			Reason:     "Insufficient FUEL balance",
		}
	} else {
		access.Street = LeagueStatus{
			Accessible: true,
			BuyinCost:  StreetBuyin,
		}
	}
	
	// Pro league
	if wallet.FuelBalance.LessThan(ProBuyin) {
		access.Pro = LeagueStatus{
			Accessible: false,
			BuyinCost:  ProBuyin,
			Reason:     "Insufficient FUEL balance",
		}
	} else {
		access.Pro = LeagueStatus{
			Accessible: true,
			BuyinCost:  ProBuyin,
		}
	}
	
	// Top Fuel league
	if wallet.FuelBalance.LessThan(TopFuelBuyin) {
		access.TopFuel = LeagueStatus{
			Accessible: false,
			BuyinCost:  TopFuelBuyin,
			Reason:     "Insufficient FUEL balance",
		}
	} else {
		access.TopFuel = LeagueStatus{
			Accessible: true,
			BuyinCost:  TopFuelBuyin,
		}
	}
	
	return access
}