package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secretKey []byte
	issuer    string
}

// Claims represents the JWT claims for our application
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	TelegramID  int64     `json:"telegram_id"`
	TokenType   string    `json:"token_type"` // "app" or "centrifugo"
	jwt.RegisteredClaims
}

// TokenType constants
const (
	TokenTypeApp        = "app"
	TokenTypeCentrifugo = "centrifugo"
)

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, issuer string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateAppToken generates a JWT token for API authentication
func (m *JWTManager) GenerateAppToken(userID uuid.UUID, telegramID int64, duration time.Duration) (string, error) {
	now := time.Now()
	
	claims := &Claims{
		UserID:     userID,
		TelegramID: telegramID,
		TokenType:  TokenTypeApp,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			Audience:  []string{"ndr-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateCentrifugoToken generates a JWT token for Centrifugo authentication
func (m *JWTManager) GenerateCentrifugoToken(userID uuid.UUID, telegramID int64, duration time.Duration) (string, error) {
	now := time.Now()
	
	claims := &Claims{
		UserID:     userID,
		TelegramID: telegramID,
		TokenType:  TokenTypeCentrifugo,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			Audience:  []string{"centrifugo"},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ValidateAppToken validates an app token and ensures it's the correct type
func (m *JWTManager) ValidateAppToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeApp {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", TokenTypeApp, claims.TokenType)
	}

	return claims, nil
}

// ValidateCentrifugoToken validates a Centrifugo token and ensures it's the correct type
func (m *JWTManager) ValidateCentrifugoToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeCentrifugo {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", TokenTypeCentrifugo, claims.TokenType)
	}

	return claims, nil
}

// RefreshToken generates a new token with the same claims but extended expiration
func (m *JWTManager) RefreshToken(tokenString string, duration time.Duration) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Generate new token with same user info but new expiration
	switch claims.TokenType {
	case TokenTypeApp:
		return m.GenerateAppToken(claims.UserID, claims.TelegramID, duration)
	case TokenTypeCentrifugo:
		return m.GenerateCentrifugoToken(claims.UserID, claims.TelegramID, duration)
	default:
		return "", fmt.Errorf("unknown token type: %s", claims.TokenType)
	}
}

// ExtractUserIDFromToken extracts user ID from a token without full validation
// Useful for logging/metrics where we don't need full security validation
func (m *JWTManager) ExtractUserIDFromToken(tokenString string) (uuid.UUID, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid token claims")
	}

	return claims.UserID, nil
}

// TokenInfo represents information about a token
type TokenInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	TelegramID  int64     `json:"telegram_id"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	IssuedAt    time.Time `json:"issued_at"`
	TokenID     string    `json:"token_id"`
}

// GetTokenInfo extracts information from a token without validating signature
func (m *JWTManager) GetTokenInfo(tokenString string) (*TokenInfo, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	info := &TokenInfo{
		UserID:     claims.UserID,
		TelegramID: claims.TelegramID,
		TokenType:  claims.TokenType,
		TokenID:    claims.ID,
	}

	if claims.ExpiresAt != nil {
		info.ExpiresAt = claims.ExpiresAt.Time
	}

	if claims.IssuedAt != nil {
		info.IssuedAt = claims.IssuedAt.Time
	}

	return info, nil
}