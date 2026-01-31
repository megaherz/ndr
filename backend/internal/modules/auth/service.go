package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"ndr/internal/auth"
	"ndr/internal/storage/postgres/models"
	"ndr/internal/storage/postgres/repository"
)

// AuthService handles authentication operations
type AuthService interface {
	// Authenticate validates Telegram initData and returns JWT tokens
	Authenticate(ctx context.Context, initData string) (*AuthResult, error)
	
	// ValidateToken validates a JWT token and returns user info
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	
	// RefreshToken generates a new access token from a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
}

// AuthResult represents the result of a successful authentication
type AuthResult struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`    // Access token expiry in seconds
	TokenType    string       `json:"token_type"`    // "Bearer"
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID       uuid.UUID `json:"user_id"`
	TelegramID   int64     `json:"telegram_id"`
	IssuedAt     int64     `json:"iat"`
	ExpiresAt    int64     `json:"exp"`
	TokenType    string    `json:"token_type"` // "access" or "refresh"
}

// authService implements AuthService
type authService struct {
	userRepo      repository.UserRepository
	walletRepo    repository.WalletRepository
	jwtUtil       *auth.JWTUtil
	botToken      string
	logger        *logrus.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	walletRepo repository.WalletRepository,
	jwtUtil *auth.JWTUtil,
	botToken string,
	logger *logrus.Logger,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		walletRepo: walletRepo,
		jwtUtil:    jwtUtil,
		botToken:   botToken,
		logger:     logger,
	}
}

// Authenticate validates Telegram initData and returns JWT tokens
func (s *authService) Authenticate(ctx context.Context, initData string) (*AuthResult, error) {
	// Validate Telegram initData
	telegramData, err := ValidateTelegramInitData(initData, s.botToken)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Invalid Telegram initData")
		return nil, fmt.Errorf("invalid telegram data: %w", err)
	}
	
	// Get or create user
	user, err := s.userRepo.GetOrCreateByTelegramID(
		ctx,
		telegramData.User.ID,
		telegramData.User.Username,
		telegramData.User.FirstName,
		telegramData.User.LastName,
	)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"telegram_id": telegramData.User.ID,
			"error":       err,
		}).Error("Failed to get or create user")
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}
	
	// Ensure user has a wallet
	err = s.ensureUserWallet(ctx, user)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Error("Failed to ensure user wallet")
		return nil, fmt.Errorf("failed to ensure user wallet: %w", err)
	}
	
	// Generate JWT tokens
	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID, telegramData.User.ID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Error("Failed to generate access token")
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	
	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID, telegramData.User.ID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Error("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"telegram_id": telegramData.User.ID,
	}).Info("User authenticated successfully")
	
	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *authService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtUtil.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	
	// Convert JWT claims to our TokenClaims struct
	tokenClaims := &TokenClaims{
		UserID:     claims.UserID,
		TelegramID: claims.TelegramID,
		IssuedAt:   claims.IssuedAt.Unix(),
		ExpiresAt:  claims.ExpiresAt.Unix(),
		TokenType:  claims.TokenType,
	}
	
	return tokenClaims, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Validate refresh token
	claims, err := s.jwtUtil.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	
	// Ensure it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}
	
	// Get user to ensure they still exist
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	
	// Generate new access token
	accessToken, err := s.jwtUtil.GenerateAccessToken(claims.UserID, claims.TelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	
	// Generate new refresh token
	newRefreshToken, err := s.jwtUtil.GenerateRefreshToken(claims.UserID, claims.TelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    3600, // 1 hour
		TokenType:    "Bearer",
	}, nil
}

// ensureUserWallet creates a wallet for the user if it doesn't exist
func (s *authService) ensureUserWallet(ctx context.Context, user *models.User) error {
	// Check if wallet exists
	wallet, err := s.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return err
	}
	
	if wallet != nil {
		return nil // Wallet already exists
	}
	
	// Create new wallet
	newWallet := &models.Wallet{
		UserID:               user.ID,
		TONBalance:           decimal.Zero,
		FuelBalance:          decimal.Zero,
		BurnBalance:          decimal.Zero,
		RookieRacesCompleted: 0,
		TONWalletAddress:     sql.NullString{},
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	
	return s.walletRepo.Create(ctx, newWallet)
}
