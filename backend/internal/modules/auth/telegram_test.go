package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBotToken = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
)

// Helper function to create a valid Telegram initData string
func createValidInitData(t testing.TB, botToken string, user *TelegramUser, authDate int64, queryID string) string {
	// Create user JSON
	userJSON := fmt.Sprintf(`{"id":%d,"first_name":"%s"`, user.ID, user.FirstName)
	if user.LastName != "" {
		userJSON += fmt.Sprintf(`,"last_name":"%s"`, user.LastName)
	}
	if user.Username != "" {
		userJSON += fmt.Sprintf(`,"username":"%s"`, user.Username)
	}
	userJSON += "}"

	// URL encode the user JSON
	encodedUser := url.QueryEscape(userJSON)

	// Create parameters
	params := map[string]string{
		"user":      encodedUser,
		"auth_date": strconv.FormatInt(authDate, 10),
	}
	if queryID != "" {
		params["query_id"] = queryID
	}

	// Create data check string
	var pairs []string
	for key, value := range params {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Create secret key
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	// Calculate hash
	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))

	// Build initData string
	var initDataPairs []string
	for key, value := range params {
		initDataPairs = append(initDataPairs, fmt.Sprintf("%s=%s", key, value))
	}
	initDataPairs = append(initDataPairs, fmt.Sprintf("hash=%s", hashHex))

	return strings.Join(initDataPairs, "&")
}

func TestValidateTelegramInitData_Success(t *testing.T) {
	user := &TelegramUser{
		ID:        123456789,
		Username:  "testuser",
		FirstName: "John",
		LastName:  "Doe",
	}

	authDate := time.Now().Unix()
	queryID := "AAHdF6IQAAAAAN0XohDhrOrc"

	initData := createValidInitData(t, testBotToken, user, authDate, queryID)

	result, err := ValidateTelegramInitData(initData, testBotToken)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, user.ID, result.User.ID)
	assert.Equal(t, user.Username, result.User.Username)
	assert.Equal(t, user.FirstName, result.User.FirstName)
	assert.Equal(t, user.LastName, result.User.LastName)
	assert.Equal(t, authDate, result.AuthDate)
	assert.Equal(t, queryID, result.QueryID)
	assert.NotEmpty(t, result.Hash)
}

func TestValidateTelegramInitData_MinimalUser(t *testing.T) {
	user := &TelegramUser{
		ID:        987654321,
		FirstName: "Jane",
		// No username or last name
	}

	authDate := time.Now().Unix()

	initData := createValidInitData(t, testBotToken, user, authDate, "")

	result, err := ValidateTelegramInitData(initData, testBotToken)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, user.ID, result.User.ID)
	assert.Equal(t, "", result.User.Username)
	assert.Equal(t, user.FirstName, result.User.FirstName)
	assert.Equal(t, "", result.User.LastName)
	assert.Equal(t, authDate, result.AuthDate)
	assert.Equal(t, "", result.QueryID)
}

func TestValidateTelegramInitData_EmptyInitData(t *testing.T) {
	result, err := ValidateTelegramInitData("", testBotToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInitData, err)
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_InvalidFormat(t *testing.T) {
	result, err := ValidateTelegramInitData("invalid%format", testBotToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse init data")
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_MissingHash(t *testing.T) {
	initData := "user=%7B%22id%22%3A123456789%7D&auth_date=1640995200"

	result, err := ValidateTelegramInitData(initData, testBotToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_InvalidHash(t *testing.T) {
	user := &TelegramUser{
		ID:        123456789,
		FirstName: "John",
	}

	authDate := time.Now().Unix()
	initData := createValidInitData(t, testBotToken, user, authDate, "")

	// Tamper with the hash
	initData = strings.Replace(initData, "hash=", "hash=invalid", 1)

	result, err := ValidateTelegramInitData(initData, testBotToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_WrongBotToken(t *testing.T) {
	user := &TelegramUser{
		ID:        123456789,
		FirstName: "John",
	}

	authDate := time.Now().Unix()
	initData := createValidInitData(t, testBotToken, user, authDate, "")

	// Use different bot token for validation
	wrongToken := "654321:XYZ-ABC9876defGhi-abc12W3v4u567fg89"

	result, err := ValidateTelegramInitData(initData, wrongToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidHash, err)
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_ExpiredData(t *testing.T) {
	user := &TelegramUser{
		ID:        123456789,
		FirstName: "John",
	}

	// Create data that's 25 hours old (expired)
	expiredAuthDate := time.Now().Unix() - (25 * 60 * 60)
	initData := createValidInitData(t, testBotToken, user, expiredAuthDate, "")

	result, err := ValidateTelegramInitData(initData, testBotToken)

	assert.Error(t, err)
	assert.Equal(t, ErrExpiredInitData, err)
	assert.Nil(t, result)
}

func TestValidateTelegramInitData_InvalidUserJSON(t *testing.T) {
	authDate := time.Now().Unix()
	invalidUserJSON := url.QueryEscape(`{"id":123456789,"first_name":"John"`) // Missing closing brace

	// We need to create a valid hash for the invalid data
	dataPairs := []string{
		fmt.Sprintf("auth_date=%d", authDate),
		fmt.Sprintf("user=%s", invalidUserJSON),
	}
	sort.Strings(dataPairs)
	dataCheckString := strings.Join(dataPairs, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(testBotToken))
	secret := secretKey.Sum(nil)

	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))

	initData := fmt.Sprintf("user=%s&auth_date=%d&hash=%s", invalidUserJSON, authDate, hashHex)

	result, err := ValidateTelegramInitData(initData, testBotToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse user data")
	assert.Nil(t, result)
}

// Test with valid hash but missing auth_date to test specific validation
func TestValidateTelegramInitData_ValidHashMissingAuthDate(t *testing.T) {
	userJSON := url.QueryEscape(`{"id":123456789,"first_name":"John"}`)

	// Create valid hash for data without auth_date
	dataPairs := []string{
		fmt.Sprintf("user=%s", userJSON),
	}
	dataCheckString := strings.Join(dataPairs, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(testBotToken))
	secret := secretKey.Sum(nil)

	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))

	initData := fmt.Sprintf("user=%s&hash=%s", userJSON, hashHex)

	result, err := ValidateTelegramInitData(initData, testBotToken)

	assert.Error(t, err)
	assert.Equal(t, ErrMissingRequiredData, err)
	assert.Nil(t, result)
}

// Test with valid hash but invalid auth_date format
func TestValidateTelegramInitData_ValidHashInvalidAuthDate(t *testing.T) {
	userJSON := url.QueryEscape(`{"id":123456789,"first_name":"John"}`)
	invalidAuthDate := "invalid"

	// Create valid hash for data with invalid auth_date
	dataPairs := []string{
		fmt.Sprintf("auth_date=%s", invalidAuthDate),
		fmt.Sprintf("user=%s", userJSON),
	}
	sort.Strings(dataPairs)
	dataCheckString := strings.Join(dataPairs, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(testBotToken))
	secret := secretKey.Sum(nil)

	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))

	initData := fmt.Sprintf("user=%s&auth_date=%s&hash=%s", userJSON, invalidAuthDate, hashHex)

	result, err := ValidateTelegramInitData(initData, testBotToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth_date")
	assert.Nil(t, result)
}

// Test with valid hash but missing user
func TestValidateTelegramInitData_ValidHashMissingUser(t *testing.T) {
	authDate := time.Now().Unix()

	// Create valid hash for data without user
	dataPairs := []string{
		fmt.Sprintf("auth_date=%d", authDate),
	}
	dataCheckString := strings.Join(dataPairs, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(testBotToken))
	secret := secretKey.Sum(nil)

	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))

	initData := fmt.Sprintf("auth_date=%d&hash=%s", authDate, hashHex)

	result, err := ValidateTelegramInitData(initData, testBotToken)

	assert.Error(t, err)
	assert.Equal(t, ErrMissingRequiredData, err)
	assert.Nil(t, result)
}

func TestParseTelegramUser_Success(t *testing.T) {
	tests := []struct {
		name     string
		userJSON string
		expected *TelegramUser
	}{
		{
			name:     "Full user data",
			userJSON: `{"id":123456789,"username":"testuser","first_name":"John","last_name":"Doe"}`,
			expected: &TelegramUser{
				ID:        123456789,
				Username:  "testuser",
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		{
			name:     "Minimal user data",
			userJSON: `{"id":987654321,"first_name":"Jane"}`,
			expected: &TelegramUser{
				ID:        987654321,
				Username:  "",
				FirstName: "Jane",
				LastName:  "",
			},
		},
		{
			name:     "User with username only",
			userJSON: `{"id":555666777,"username":"cooluser","first_name":"Bob"}`,
			expected: &TelegramUser{
				ID:        555666777,
				Username:  "cooluser",
				FirstName: "Bob",
				LastName:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// URL encode the JSON as Telegram does
			encodedJSON := url.QueryEscape(tt.userJSON)

			result, err := parseTelegramUser(encodedJSON)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Username, result.Username)
			assert.Equal(t, tt.expected.FirstName, result.FirstName)
			assert.Equal(t, tt.expected.LastName, result.LastName)
		})
	}
}

func TestParseTelegramUser_InvalidURLEncoding(t *testing.T) {
	result, err := parseTelegramUser("%zzinvalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to URL decode user JSON")
	assert.Nil(t, result)
}

func TestParseTelegramUser_InvalidJSON(t *testing.T) {
	invalidJSON := url.QueryEscape(`{"id":123,"first_name":"John"`) // Missing closing brace

	result, err := parseTelegramUser(invalidJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse user JSON")
	assert.Nil(t, result)
}

func TestParseTelegramUser_MissingID(t *testing.T) {
	userJSON := url.QueryEscape(`{"first_name":"John"}`)

	result, err := parseTelegramUser(userJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing or invalid user ID")
	assert.Nil(t, result)
}

func TestParseTelegramUser_ZeroID(t *testing.T) {
	userJSON := url.QueryEscape(`{"id":0,"first_name":"John"}`)

	result, err := parseTelegramUser(userJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing or invalid user ID")
	assert.Nil(t, result)
}

func TestParseTelegramUser_MissingFirstName(t *testing.T) {
	userJSON := url.QueryEscape(`{"id":123456789}`)

	result, err := parseTelegramUser(userJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing first_name")
	assert.Nil(t, result)
}

func TestParseTelegramUser_EmptyFirstName(t *testing.T) {
	userJSON := url.QueryEscape(`{"id":123456789,"first_name":""}`)

	result, err := parseTelegramUser(userJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing first_name")
	assert.Nil(t, result)
}

func TestParseTelegramUser_WithSpecialCharacters(t *testing.T) {
	userJSON := url.QueryEscape(`{"id":123456789,"first_name":"José","last_name":"García-López","username":"jose_garcia"}`)

	result, err := parseTelegramUser(userJSON)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(123456789), result.ID)
	assert.Equal(t, "José", result.FirstName)
	assert.Equal(t, "García-López", result.LastName)
	assert.Equal(t, "jose_garcia", result.Username)
}

// Benchmark tests
func BenchmarkValidateTelegramInitData(b *testing.B) {
	user := &TelegramUser{
		ID:        123456789,
		Username:  "testuser",
		FirstName: "John",
		LastName:  "Doe",
	}

	authDate := time.Now().Unix()
	initData := createValidInitData(b, testBotToken, user, authDate, "test_query_id")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateTelegramInitData(initData, testBotToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseTelegramUser(b *testing.B) {
	userJSON := url.QueryEscape(`{"id":123456789,"username":"testuser","first_name":"John","last_name":"Doe"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseTelegramUser(userJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test edge cases and security
func TestValidateTelegramInitData_EdgeCases(t *testing.T) {
	t.Run("Very long username", func(t *testing.T) {
		user := &TelegramUser{
			ID:        123456789,
			Username:  strings.Repeat("a", 100), // Very long username
			FirstName: "John",
		}

		authDate := time.Now().Unix()
		initData := createValidInitData(t, testBotToken, user, authDate, "")

		result, err := ValidateTelegramInitData(initData, testBotToken)
		require.NoError(t, err)
		assert.Equal(t, user.Username, result.User.Username)
	})

	t.Run("Unicode characters in names", func(t *testing.T) {
		user := &TelegramUser{
			ID:        123456789,
			FirstName: "🎮 Gamer",
			LastName:  "测试用户",
			Username:  "user_тест",
		}

		authDate := time.Now().Unix()
		initData := createValidInitData(t, testBotToken, user, authDate, "")

		result, err := ValidateTelegramInitData(initData, testBotToken)
		require.NoError(t, err)
		assert.Equal(t, user.FirstName, result.User.FirstName)
		assert.Equal(t, user.LastName, result.User.LastName)
		assert.Equal(t, user.Username, result.User.Username)
	})

	t.Run("Maximum valid age (23h 59m)", func(t *testing.T) {
		user := &TelegramUser{
			ID:        123456789,
			FirstName: "John",
		}

		// Create data that's just under 24 hours old
		authDate := time.Now().Unix() - (23*60*60 + 59*60) // 23h 59m ago
		initData := createValidInitData(t, testBotToken, user, authDate, "")

		result, err := ValidateTelegramInitData(initData, testBotToken)
		require.NoError(t, err)
		assert.Equal(t, authDate, result.AuthDate)
	})

	t.Run("Just expired (24h 1m)", func(t *testing.T) {
		user := &TelegramUser{
			ID:        123456789,
			FirstName: "John",
		}

		// Create data that's just over 24 hours old
		authDate := time.Now().Unix() - (24*60*60 + 60) // 24h 1m ago
		initData := createValidInitData(t, testBotToken, user, authDate, "")

		result, err := ValidateTelegramInitData(initData, testBotToken)
		assert.Error(t, err)
		assert.Equal(t, ErrExpiredInitData, err)
		assert.Nil(t, result)
	})
}

// Test concurrent access (important for web server)
func TestValidateTelegramInitData_Concurrent(t *testing.T) {
	user := &TelegramUser{
		ID:        123456789,
		FirstName: "John",
	}

	authDate := time.Now().Unix()
	initData := createValidInitData(t, testBotToken, user, authDate, "")

	// Run validation concurrently
	const numGoroutines = 100
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := ValidateTelegramInitData(initData, testBotToken)
			results <- err
		}()
	}

	// Check all results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent validation should succeed")
	}
}

// Test error types
func TestErrorTypes(t *testing.T) {
	assert.Equal(t, "invalid telegram init data", ErrInvalidInitData.Error())
	assert.Equal(t, "telegram init data expired", ErrExpiredInitData.Error())
	assert.Equal(t, "invalid telegram init data hash", ErrInvalidHash.Error())
	assert.Equal(t, "missing required telegram data", ErrMissingRequiredData.Error())
}
