package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

	// Remove hash from values for validation
	values.Del("hash")

	// Create data check string
	var pairs []string
	for key, vals := range values {
		if len(vals) > 0 {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, vals[0]))
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

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
// This is a simplified parser for the required fields
func parseTelegramUser(userJSON string) (*TelegramUser, error) {
	// Decode URL-encoded JSON
	decoded, err := url.QueryUnescape(userJSON)
	if err != nil {
		return nil, err
	}

	// Simple JSON parsing for required fields
	user := &TelegramUser{}

	// Extract ID
	if idStart := strings.Index(decoded, `"id":`); idStart != -1 {
		idStart += 5
		idEnd := strings.IndexAny(decoded[idStart:], ",}")
		if idEnd == -1 {
			return nil, errors.New("invalid user JSON format")
		}
		idStr := decoded[idStart : idStart+idEnd]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
		user.ID = id
	} else {
		return nil, errors.New("missing user ID")
	}

	// Extract first_name
	if nameStart := strings.Index(decoded, `"first_name":"`); nameStart != -1 {
		nameStart += 14
		nameEnd := strings.Index(decoded[nameStart:], `"`)
		if nameEnd == -1 {
			return nil, errors.New("invalid first_name format")
		}
		user.FirstName = decoded[nameStart : nameStart+nameEnd]
	} else {
		return nil, errors.New("missing first_name")
	}

	// Extract last_name (optional)
	if nameStart := strings.Index(decoded, `"last_name":"`); nameStart != -1 {
		nameStart += 13
		nameEnd := strings.Index(decoded[nameStart:], `"`)
		if nameEnd != -1 {
			user.LastName = decoded[nameStart : nameStart+nameEnd]
		}
	}

	// Extract username (optional)
	if usernameStart := strings.Index(decoded, `"username":"`); usernameStart != -1 {
		usernameStart += 12
		usernameEnd := strings.Index(decoded[usernameStart:], `"`)
		if usernameEnd != -1 {
			user.Username = decoded[usernameStart : usernameStart+usernameEnd]
		}
	}

	return user, nil
}