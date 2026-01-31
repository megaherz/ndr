package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/megaherz/ndr/internal/storage/postgres/models"
)

type UserRepositoryIntegrationTestSuite struct {
	suite.Suite
	dbHelper   *TestDBHelper
	repository UserRepository
}

func TestUserRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryIntegrationTestSuite))
}

func (suite *UserRepositoryIntegrationTestSuite) SetupSuite() {
	// Setup database helper
	suite.dbHelper = NewTestDBHelper(suite.T())
	suite.dbHelper.SetupDatabase()
	
	// Create repository instance
	suite.repository = NewUserRepository(suite.dbHelper.DB)
}

func (suite *UserRepositoryIntegrationTestSuite) TearDownSuite() {
	suite.dbHelper.TeardownDatabase()
}

func (suite *UserRepositoryIntegrationTestSuite) SetupTest() {
	// Clean up users table before each test
	suite.dbHelper.CleanupTables("users")
}


func (suite *UserRepositoryIntegrationTestSuite) TestCreate() {
	ctx := context.Background()
	
	user := &models.User{
		ID:                uuid.New(),
		TelegramID:        123456789,
		TelegramUsername:  stringPtr("testuser"),
		TelegramFirstName: "Test",
		TelegramLastName:  stringPtr("User"),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	
	err := suite.repository.Create(ctx, user)
	require.NoError(suite.T(), err)
	
	// Verify user was created
	var count int
	err = suite.dbHelper.DB.Get(&count, "SELECT COUNT(*) FROM users WHERE id = $1", user.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

func (suite *UserRepositoryIntegrationTestSuite) TestCreate_DuplicateTelegramID() {
	ctx := context.Background()
	
	telegramID := int64(123456789)
	
	// Create first user
	user1 := &models.User{
		ID:                uuid.New(),
		TelegramID:        telegramID,
		TelegramFirstName: "First",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	
	err := suite.repository.Create(ctx, user1)
	require.NoError(suite.T(), err)
	
	// Try to create second user with same Telegram ID
	user2 := &models.User{
		ID:                uuid.New(),
		TelegramID:        telegramID,
		TelegramFirstName: "Second",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	
	err = suite.repository.Create(ctx, user2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate key")
}

func (suite *UserRepositoryIntegrationTestSuite) TestGetByID() {
	ctx := context.Background()
	
	// Create a user first
	now := time.Now().UTC().Truncate(time.Microsecond) // Use UTC and truncate for comparison
	originalUser := &models.User{
		ID:                uuid.New(),
		TelegramID:        123456789,
		TelegramUsername:  stringPtr("testuser"),
		TelegramFirstName: "Test",
		TelegramLastName:  stringPtr("User"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	
	err := suite.repository.Create(ctx, originalUser)
	require.NoError(suite.T(), err)
	
	// Retrieve the user
	retrievedUser, err := suite.repository.GetByID(ctx, originalUser.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedUser)
	
	// Compare all fields
	assert.Equal(suite.T(), originalUser.ID, retrievedUser.ID)
	assert.Equal(suite.T(), originalUser.TelegramID, retrievedUser.TelegramID)
	assert.Equal(suite.T(), originalUser.TelegramUsername, retrievedUser.TelegramUsername)
	assert.Equal(suite.T(), originalUser.TelegramFirstName, retrievedUser.TelegramFirstName)
	assert.Equal(suite.T(), originalUser.TelegramLastName, retrievedUser.TelegramLastName)
	
	// Time comparison with tolerance
	assert.WithinDuration(suite.T(), originalUser.CreatedAt, retrievedUser.CreatedAt, time.Second)
	assert.WithinDuration(suite.T(), originalUser.UpdatedAt, retrievedUser.UpdatedAt, time.Second)
}

func (suite *UserRepositoryIntegrationTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()
	
	nonExistentID := uuid.New()
	user, err := suite.repository.GetByID(ctx, nonExistentID)
	
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), user)
}

func (suite *UserRepositoryIntegrationTestSuite) TestGetByTelegramID() {
	ctx := context.Background()
	
	telegramID := int64(987654321)
	
	// Create a user
	now := time.Now().UTC().Truncate(time.Microsecond)
	originalUser := &models.User{
		ID:                uuid.New(),
		TelegramID:        telegramID,
		TelegramUsername:  stringPtr("telegram_user"),
		TelegramFirstName: "Telegram",
		TelegramLastName:  stringPtr("User"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	
	err := suite.repository.Create(ctx, originalUser)
	require.NoError(suite.T(), err)
	
	// Retrieve by Telegram ID
	retrievedUser, err := suite.repository.GetByTelegramID(ctx, telegramID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedUser)
	
	assert.Equal(suite.T(), originalUser.ID, retrievedUser.ID)
	assert.Equal(suite.T(), originalUser.TelegramID, retrievedUser.TelegramID)
}

func (suite *UserRepositoryIntegrationTestSuite) TestGetByTelegramID_NotFound() {
	ctx := context.Background()
	
	user, err := suite.repository.GetByTelegramID(ctx, 999999999)
	
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), user)
}

func (suite *UserRepositoryIntegrationTestSuite) TestUpdateTelegramInfo() {
	ctx := context.Background()
	
	// Create a user
	user := &models.User{
		ID:                uuid.New(),
		TelegramID:        123456789,
		TelegramUsername:  stringPtr("oldusername"),
		TelegramFirstName: "Old",
		TelegramLastName:  stringPtr("Name"),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	
	err := suite.repository.Create(ctx, user)
	require.NoError(suite.T(), err)
	
	// Add a small delay to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)
	
	// Update Telegram info
	newUsername := "newusername"
	newFirstName := "New"
	newLastName := "UpdatedName"
	
	err = suite.repository.UpdateTelegramInfo(ctx, user.ID, newUsername, newFirstName, newLastName)
	require.NoError(suite.T(), err)
	
	// Verify update
	updatedUser, err := suite.repository.GetByID(ctx, user.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), updatedUser)
	
	assert.Equal(suite.T(), newUsername, *updatedUser.TelegramUsername)
	assert.Equal(suite.T(), newFirstName, updatedUser.TelegramFirstName)
	assert.Equal(suite.T(), newLastName, *updatedUser.TelegramLastName)
	
	// Verify updated_at was changed
	assert.True(suite.T(), updatedUser.UpdatedAt.After(user.UpdatedAt))
}

func (suite *UserRepositoryIntegrationTestSuite) TestGetOrCreateByTelegramID_ExistingUser() {
	ctx := context.Background()
	
	telegramID := int64(555666777)
	
	// Create an existing user
	existingUser := &models.User{
		ID:                uuid.New(),
		TelegramID:        telegramID,
		TelegramUsername:  stringPtr("existing"),
		TelegramFirstName: "Existing",
		TelegramLastName:  stringPtr("User"),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	
	err := suite.repository.Create(ctx, existingUser)
	require.NoError(suite.T(), err)
	
	// Add a small delay to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)
	
	// Call GetOrCreateByTelegramID with updated info
	user, err := suite.repository.GetOrCreateByTelegramID(
		ctx, 
		telegramID, 
		"updated_username", 
		"Updated", 
		"Name",
	)
	
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), user)
	
	// Should return the same user ID but with updated info
	assert.Equal(suite.T(), existingUser.ID, user.ID)
	assert.Equal(suite.T(), "updated_username", *user.TelegramUsername)
	assert.Equal(suite.T(), "Updated", user.TelegramFirstName)
	assert.Equal(suite.T(), "Name", *user.TelegramLastName)
}

func (suite *UserRepositoryIntegrationTestSuite) TestGetOrCreateByTelegramID_NewUser() {
	ctx := context.Background()
	
	telegramID := int64(888999000)
	
	// Call GetOrCreateByTelegramID for non-existing user
	user, err := suite.repository.GetOrCreateByTelegramID(
		ctx,
		telegramID,
		"new_user",
		"New",
		"User",
	)
	
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), user)
	
	// Verify it's a new user
	assert.Equal(suite.T(), telegramID, user.TelegramID)
	assert.Equal(suite.T(), "new_user", *user.TelegramUsername)
	assert.Equal(suite.T(), "New", user.TelegramFirstName)
	assert.Equal(suite.T(), "User", *user.TelegramLastName)
	
	// Verify user was actually created in database
	dbUser, err := suite.repository.GetByTelegramID(ctx, telegramID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), dbUser)
	assert.Equal(suite.T(), user.ID, dbUser.ID)
}

func (suite *UserRepositoryIntegrationTestSuite) TestGetOrCreateByTelegramID_EmptyOptionalFields() {
	ctx := context.Background()
	
	telegramID := int64(111222333)
	
	// Call with empty username and last name
	user, err := suite.repository.GetOrCreateByTelegramID(
		ctx,
		telegramID,
		"", // empty username
		"FirstOnly",
		"", // empty last name
	)
	
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), user)
	
	// Verify optional fields are nil when empty
	assert.Nil(suite.T(), user.TelegramUsername)
	assert.Equal(suite.T(), "FirstOnly", user.TelegramFirstName)
	assert.Nil(suite.T(), user.TelegramLastName)
}

func (suite *UserRepositoryIntegrationTestSuite) TestList() {
	ctx := context.Background()
	
	// Create multiple users
	now := time.Now().UTC()
	users := []*models.User{
		{
			ID:                uuid.New(),
			TelegramID:        111,
			TelegramFirstName: "User1",
			CreatedAt:         now.Add(-3 * time.Hour),
			UpdatedAt:         now.Add(-3 * time.Hour),
		},
		{
			ID:                uuid.New(),
			TelegramID:        222,
			TelegramFirstName: "User2",
			CreatedAt:         now.Add(-2 * time.Hour),
			UpdatedAt:         now.Add(-2 * time.Hour),
		},
		{
			ID:                uuid.New(),
			TelegramID:        333,
			TelegramFirstName: "User3",
			CreatedAt:         now.Add(-1 * time.Hour),
			UpdatedAt:         now.Add(-1 * time.Hour),
		},
	}
	
	for _, user := range users {
		err := suite.repository.Create(ctx, user)
		require.NoError(suite.T(), err)
	}
	
	// Test pagination
	retrievedUsers, err := suite.repository.List(ctx, 2, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrievedUsers, 2)
	
	// Should be ordered by created_at DESC (newest first)
	assert.Equal(suite.T(), "User3", retrievedUsers[0].TelegramFirstName)
	assert.Equal(suite.T(), "User2", retrievedUsers[1].TelegramFirstName)
	
	// Test second page
	retrievedUsers, err = suite.repository.List(ctx, 2, 2)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrievedUsers, 1)
	assert.Equal(suite.T(), "User1", retrievedUsers[0].TelegramFirstName)
}

func (suite *UserRepositoryIntegrationTestSuite) TestCount() {
	ctx := context.Background()
	
	// Initially should be 0
	count, err := suite.repository.Count(ctx)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
	
	// Create some users
	for i := 0; i < 5; i++ {
		user := &models.User{
			ID:                uuid.New(),
			TelegramID:        int64(100 + i),
			TelegramFirstName: fmt.Sprintf("User%d", i),
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}
		
		err := suite.repository.Create(ctx, user)
		require.NoError(suite.T(), err)
	}
	
	// Count should be 5
	count, err = suite.repository.Count(ctx)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(5), count)
}