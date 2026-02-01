package auth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
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
// using the official telegram-mini-apps/init-data-golang package
func ValidateTelegramInitData(initDataRaw, botToken string) (*TelegramInitData, error) {
	if initDataRaw == "" {
		return nil, ErrInvalidInitData
	}

	// Extract bot ID from bot token (format: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	parts := strings.Split(botToken, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid bot token format")
	}

	botID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bot ID in token: %w", err)
	}

	// Set expiration time to 24 hours
	expIn := 24 * time.Hour

	// Try both validation methods since we have both hash and signature fields

	// First, try regular bot token validation
	err = initdata.Validate(initDataRaw, botToken, expIn)
	if err != nil {
		// Try third-party validation with bot ID
		err = initdata.ValidateThirdParty(initDataRaw, botID, expIn)
		if err != nil {
			return nil, fmt.Errorf("validation failed with both methods: %w", err)
		}
	}

	// Parse the validated initData
	parsedData, err := initdata.Parse(initDataRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse init data: %w", err)
	}

	// Convert to our internal format
	if parsedData.User.ID == 0 {
		return nil, ErrMissingRequiredData
	}

	user := &TelegramUser{
		ID:        parsedData.User.ID,
		Username:  parsedData.User.Username,
		FirstName: parsedData.User.FirstName,
		LastName:  parsedData.User.LastName,
		PhotoURL:  parsedData.User.PhotoURL,
	}

	return &TelegramInitData{
		User:     user,
		AuthDate: parsedData.AuthDate().Unix(),
		Hash:     parsedData.Hash,
		QueryID:  parsedData.QueryID,
	}, nil
}
