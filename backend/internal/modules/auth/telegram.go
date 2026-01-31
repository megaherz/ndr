package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidInitData     = errors.New("invalid telegram init data")
	ErrExpiredInitData     = errors.New("telegram init data expired")
	ErrInvalidHash         = errors.New("invalid telegram init data hash")
	ErrMissingRequiredData = errors.New("missing required telegram data")
)

// TelegramUser represents the user data from Telegram initData
type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
}

// TelegramInitData represents the parsed Telegram Web App initData
type TelegramInitData struct {
	User     *TelegramUser `json:"user"`
	AuthDate int64         `json:"auth_date"`
	Hash     string        `json:"hash"`
	QueryID  string        `json:"query_id,omitempty"`
}

// ValidateTelegramInitData validates the Telegram Web App initData
// according to the official Telegram documentation
func ValidateTelegramInitData(initData, botToken string) (*TelegramInitData, error) {
	if initData == "" {
		return nil, ErrInvalidInitData
	}

	// Parse the URL-encoded initData
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse init data: %w", err)
	}

	// Extract hash
	hash := values.Get("hash")
	if hash == "" {
		return nil, ErrInvalidHash
	}

	// Create data check string using original encoded values
	// We need to reconstruct from the original initData string to preserve encoding
	pairs := strings.Split(initData, "&")
	var dataPairs []string
	for _, pair := range pairs {
		if !strings.HasPrefix(pair, "hash=") {
			dataPairs = append(dataPairs, pair)
		}
	}
	sort.Strings(dataPairs)
	dataCheckString := strings.Join(dataPairs, "\n")

	// Create secret key
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	// Calculate expected hash
	expectedHash := hmac.New(sha256.New, secret)
	expectedHash.Write([]byte(dataCheckString))
	expectedHashHex := hex.EncodeToString(expectedHash.Sum(nil))

	// Verify hash
	if !hmac.Equal([]byte(hash), []byte(expectedHashHex)) {
		return nil, ErrInvalidHash
	}

	// Parse auth_date
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return nil, ErrMissingRequiredData
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date: %w", err)
	}

	// Check if data is not too old (24 hours)
	if time.Now().Unix()-authDate > 86400 {
		return nil, ErrExpiredInitData
	}

	// Parse user data
	userStr := values.Get("user")
	if userStr == "" {
		return nil, ErrMissingRequiredData
	}

	// Parse user JSON manually (simple parsing for required fields)
	user, err := parseTelegramUser(userStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &TelegramInitData{
		User:     user,
		AuthDate: authDate,
		Hash:     hash,
		QueryID:  values.Get("query_id"),
	}, nil
}

// parseTelegramUser parses the user JSON string from Telegram initData
func parseTelegramUser(userJSON string) (*TelegramUser, error) {
	// Decode URL-encoded JSON
	decoded, err := url.QueryUnescape(userJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to URL decode user JSON: %w", err)
	}

	// Parse JSON properly
	var user TelegramUser
	if err := json.Unmarshal([]byte(decoded), &user); err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	// Validate required fields
	if user.ID == 0 {
		return nil, errors.New("missing or invalid user ID")
	}
	if user.FirstName == "" {
		return nil, errors.New("missing first_name")
	}

	return &user, nil
}