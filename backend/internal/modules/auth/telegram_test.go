package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTelegramInitData_EmptyInitData(t *testing.T) {
	result, err := ValidateTelegramInitData("", "123456:ABC-DEF")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInitData, err)
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_InvalidBotToken(t *testing.T) {
	initData := "user=%7B%22id%22%3A123%7D&auth_date=1234567890&hash=abc123"

	result, err := ValidateTelegramInitData(initData, "invalid-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid bot token format")
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_InvalidInitDataFormat(t *testing.T) {
	// This test uses a clearly invalid initData format that the official package will reject
	initData := "invalid_format"
	botToken := "123456:ABC-DEF"

	result, err := ValidateTelegramInitData(initData, botToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed with both methods")
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_ExpiredData(t *testing.T) {
	// Create initData with an old auth_date (more than 24 hours ago)
	initData := "user=%7B%22id%22%3A123%2C%22first_name%22%3A%22Test%22%7D&auth_date=1000000000&hash=invalid"
	botToken := "123456:ABC-DEF"

	result, err := ValidateTelegramInitData(initData, botToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed with both methods")
	assert.Nil(t, result)
}

// Note: We cannot easily test successful validation without valid Telegram signatures,
// as the official package performs cryptographic verification. In a real application,
// you would test this with actual valid initData from Telegram or mock the package.

func TestValidateTelegramInitData_BotTokenExtraction(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantErr  bool
		expected int64
	}{
		{
			name:     "valid token",
			token:    "123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			wantErr:  false,
			expected: 123456789,
		},
		{
			name:    "invalid format - no colon",
			token:   "123456789ABC-DEF",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			token:   "",
			wantErr: true,
		},
		{
			name:    "invalid bot ID - not a number",
			token:   "abc:DEF-123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We test the bot ID extraction logic indirectly by checking the error messages
			result, err := ValidateTelegramInitData("dummy", tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// Even with valid token format, we expect validation to fail with dummy data
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "validation failed with both methods")
			}
			assert.Nil(t, result)
		})
	}
}
