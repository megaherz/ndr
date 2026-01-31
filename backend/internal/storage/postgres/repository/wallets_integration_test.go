package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/megaherz/ndr/internal/storage/postgres/models"
)

type WalletRepositoryIntegrationTestSuite struct {
	suite.Suite
	dbHelper   *TestDBHelper
	walletRepo WalletRepository
	userRepo   UserRepository
	testUserID uuid.UUID
}

func TestWalletRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(WalletRepositoryIntegrationTestSuite))
}

func (suite *WalletRepositoryIntegrationTestSuite) SetupSuite() {
	// Setup database helper
	suite.dbHelper = NewTestDBHelper(suite.T())
	suite.dbHelper.SetupDatabase()
	
	// Create repository instances
	suite.walletRepo = NewWalletRepository(suite.dbHelper.DB)
	suite.userRepo = NewUserRepository(suite.dbHelper.DB)
}

func (suite *WalletRepositoryIntegrationTestSuite) TearDownSuite() {
	suite.dbHelper.TeardownDatabase()
}

func (suite *WalletRepositoryIntegrationTestSuite) SetupTest() {
	// Clean up tables before each test
	suite.dbHelper.CleanupTables("wallets", "users")
	
	// Create a test user for wallet operations
	suite.testUserID = uuid.New()
	testUser := &models.User{
		ID:                suite.testUserID,
		TelegramID:        123456789,
		TelegramFirstName: "Test",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	
	err := suite.userRepo.Create(context.Background(), testUser)
	require.NoError(suite.T(), err)
}


func (suite *WalletRepositoryIntegrationTestSuite) TestCreate() {
	ctx := context.Background()
	
	wallet := &models.Wallet{
		UserID:               suite.testUserID,
		TonBalance:           decimal.NewFromFloat(100.50),
		FuelBalance:          decimal.NewFromFloat(250.75),
		BurnBalance:          decimal.NewFromFloat(0.00),
		RookieRacesCompleted: 0,
		TonWalletAddress:     stringPtr("EQD4FPq-PRDieyQKkizFTRtSDyucUIqrj0v_zXJmqaDp6_0t"),
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, wallet)
	require.NoError(suite.T(), err)
	
	// Verify wallet was created
	var count int
	err = suite.dbHelper.DB.Get(&count, "SELECT COUNT(*) FROM wallets WHERE user_id = $1", suite.testUserID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

func (suite *WalletRepositoryIntegrationTestSuite) TestCreate_DuplicateUserID() {
	ctx := context.Background()
	
	// Create first wallet
	wallet1 := &models.Wallet{
		UserID:      suite.testUserID,
		TonBalance:  decimal.NewFromFloat(100.00),
		FuelBalance: decimal.NewFromFloat(200.00),
		BurnBalance: decimal.NewFromFloat(0.00),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, wallet1)
	require.NoError(suite.T(), err)
	
	// Try to create second wallet for same user
	wallet2 := &models.Wallet{
		UserID:      suite.testUserID,
		TonBalance:  decimal.NewFromFloat(50.00),
		FuelBalance: decimal.NewFromFloat(100.00),
		BurnBalance: decimal.NewFromFloat(0.00),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	
	err = suite.walletRepo.Create(ctx, wallet2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate key")
}

func (suite *WalletRepositoryIntegrationTestSuite) TestCreate_InvalidUserID() {
	ctx := context.Background()
	
	nonExistentUserID := uuid.New()
	wallet := &models.Wallet{
		UserID:      nonExistentUserID,
		TonBalance:  decimal.NewFromFloat(100.00),
		FuelBalance: decimal.NewFromFloat(200.00),
		BurnBalance: decimal.NewFromFloat(0.00),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, wallet)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "foreign key")
}

func (suite *WalletRepositoryIntegrationTestSuite) TestGetByUserID() {
	ctx := context.Background()
	
	// Create a wallet first
	originalWallet := &models.Wallet{
		UserID:               suite.testUserID,
		TonBalance:           decimal.NewFromFloat(123.45),
		FuelBalance:          decimal.NewFromFloat(678.90),
		BurnBalance:          decimal.NewFromFloat(12.34),
		RookieRacesCompleted: 2,
		TonWalletAddress:     stringPtr("EQD4FPq-PRDieyQKkizFTRtSDyucUIqrj0v_zXJmqaDp6_0t"),
		CreatedAt:            time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt:            time.Now().UTC().Truncate(time.Microsecond),
	}
	
	err := suite.walletRepo.Create(ctx, originalWallet)
	require.NoError(suite.T(), err)
	
	// Retrieve the wallet
	retrievedWallet, err := suite.walletRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedWallet)
	
	// Compare all fields
	assert.Equal(suite.T(), originalWallet.UserID, retrievedWallet.UserID)
	assert.True(suite.T(), originalWallet.TonBalance.Equal(retrievedWallet.TonBalance))
	assert.True(suite.T(), originalWallet.FuelBalance.Equal(retrievedWallet.FuelBalance))
	assert.True(suite.T(), originalWallet.BurnBalance.Equal(retrievedWallet.BurnBalance))
	assert.Equal(suite.T(), originalWallet.RookieRacesCompleted, retrievedWallet.RookieRacesCompleted)
	assert.Equal(suite.T(), originalWallet.TonWalletAddress, retrievedWallet.TonWalletAddress)
	
	// Time comparison with tolerance
	assert.WithinDuration(suite.T(), originalWallet.CreatedAt, retrievedWallet.CreatedAt, time.Second)
	assert.WithinDuration(suite.T(), originalWallet.UpdatedAt, retrievedWallet.UpdatedAt, time.Second)
}

func (suite *WalletRepositoryIntegrationTestSuite) TestGetByUserID_NotFound() {
	ctx := context.Background()
	
	nonExistentUserID := uuid.New()
	wallet, err := suite.walletRepo.GetByUserID(ctx, nonExistentUserID)
	
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), wallet)
}

func (suite *WalletRepositoryIntegrationTestSuite) TestUpdateBalances() {
	ctx := context.Background()
	
	// Create initial wallet
	initialWallet := &models.Wallet{
		UserID:      suite.testUserID,
		TonBalance:  decimal.NewFromFloat(100.00),
		FuelBalance: decimal.NewFromFloat(200.00),
		BurnBalance: decimal.NewFromFloat(50.00),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, initialWallet)
	require.NoError(suite.T(), err)
	
	// Add a small delay to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)
	
	// Update balances
	tonDelta := decimal.NewFromFloat(25.50)
	fuelDelta := decimal.NewFromFloat(-75.25)
	burnDelta := decimal.NewFromFloat(10.75)
	
	err = suite.walletRepo.UpdateBalances(ctx, suite.testUserID, tonDelta, fuelDelta, burnDelta)
	require.NoError(suite.T(), err)
	
	// Verify updated balances
	updatedWallet, err := suite.walletRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), updatedWallet)
	
	expectedTonBalance := decimal.NewFromFloat(125.50) // 100.00 + 25.50
	expectedFuelBalance := decimal.NewFromFloat(124.75) // 200.00 - 75.25
	expectedBurnBalance := decimal.NewFromFloat(60.75)  // 50.00 + 10.75
	
	assert.True(suite.T(), expectedTonBalance.Equal(updatedWallet.TonBalance))
	assert.True(suite.T(), expectedFuelBalance.Equal(updatedWallet.FuelBalance))
	assert.True(suite.T(), expectedBurnBalance.Equal(updatedWallet.BurnBalance))
	
	// Verify updated_at was changed
	assert.True(suite.T(), updatedWallet.UpdatedAt.After(initialWallet.UpdatedAt))
}

func (suite *WalletRepositoryIntegrationTestSuite) TestUpdateBalances_NegativeResult() {
	ctx := context.Background()
	
	// Create wallet with small balance
	initialWallet := &models.Wallet{
		UserID:      suite.testUserID,
		TonBalance:  decimal.NewFromFloat(10.00),
		FuelBalance: decimal.NewFromFloat(20.00),
		BurnBalance: decimal.NewFromFloat(5.00),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, initialWallet)
	require.NoError(suite.T(), err)
	
	// Try to subtract more than available (should violate check constraint)
	tonDelta := decimal.NewFromFloat(-15.00) // Would result in -5.00
	fuelDelta := decimal.NewFromFloat(0)
	burnDelta := decimal.NewFromFloat(0)
	
	err = suite.walletRepo.UpdateBalances(ctx, suite.testUserID, tonDelta, fuelDelta, burnDelta)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "check constraint")
}

func (suite *WalletRepositoryIntegrationTestSuite) TestUpdateBalances_NonExistentUser() {
	ctx := context.Background()
	
	nonExistentUserID := uuid.New()
	tonDelta := decimal.NewFromFloat(10.00)
	fuelDelta := decimal.NewFromFloat(20.00)
	burnDelta := decimal.NewFromFloat(5.00)
	
	err := suite.walletRepo.UpdateBalances(ctx, nonExistentUserID, tonDelta, fuelDelta, burnDelta)
	require.NoError(suite.T(), err) // UPDATE with no matching rows doesn't error
	
	// Verify no wallet was created
	wallet, err := suite.walletRepo.GetByUserID(ctx, nonExistentUserID)
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), wallet)
}

func (suite *WalletRepositoryIntegrationTestSuite) TestIncrementRookieRaces() {
	ctx := context.Background()
	
	// Create wallet with initial rookie races
	initialWallet := &models.Wallet{
		UserID:               suite.testUserID,
		TonBalance:           decimal.NewFromFloat(100.00),
		FuelBalance:          decimal.NewFromFloat(200.00),
		BurnBalance:          decimal.NewFromFloat(0.00),
		RookieRacesCompleted: 1,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, initialWallet)
	require.NoError(suite.T(), err)
	
	// Add a small delay to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)
	
	// Increment rookie races
	err = suite.walletRepo.IncrementRookieRaces(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	
	// Verify increment
	updatedWallet, err := suite.walletRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), updatedWallet)
	
	assert.Equal(suite.T(), 2, updatedWallet.RookieRacesCompleted)
	
	// Verify updated_at was changed
	assert.True(suite.T(), updatedWallet.UpdatedAt.After(initialWallet.UpdatedAt))
}

func (suite *WalletRepositoryIntegrationTestSuite) TestIncrementRookieRaces_MaxLimit() {
	ctx := context.Background()
	
	// Create wallet at max rookie races (3)
	initialWallet := &models.Wallet{
		UserID:               suite.testUserID,
		TonBalance:           decimal.NewFromFloat(100.00),
		FuelBalance:          decimal.NewFromFloat(200.00),
		BurnBalance:          decimal.NewFromFloat(0.00),
		RookieRacesCompleted: 3,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, initialWallet)
	require.NoError(suite.T(), err)
	
	// Try to increment beyond max (should violate check constraint)
	err = suite.walletRepo.IncrementRookieRaces(ctx, suite.testUserID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "check constraint")
}

func (suite *WalletRepositoryIntegrationTestSuite) TestSetTONWalletAddress() {
	ctx := context.Background()
	
	// Create wallet without TON address
	initialWallet := &models.Wallet{
		UserID:           suite.testUserID,
		TonBalance:       decimal.NewFromFloat(100.00),
		FuelBalance:      decimal.NewFromFloat(200.00),
		BurnBalance:      decimal.NewFromFloat(0.00),
		TonWalletAddress: nil,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, initialWallet)
	require.NoError(suite.T(), err)
	
	// Add a small delay to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)
	
	// Set TON wallet address
	newAddress := "EQD4FPq-PRDieyQKkizFTRtSDyucUIqrj0v_zXJmqaDp6_0t"
	err = suite.walletRepo.SetTONWalletAddress(ctx, suite.testUserID, newAddress)
	require.NoError(suite.T(), err)
	
	// Verify address was set
	updatedWallet, err := suite.walletRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), updatedWallet)
	
	require.NotNil(suite.T(), updatedWallet.TonWalletAddress)
	assert.Equal(suite.T(), newAddress, *updatedWallet.TonWalletAddress)
	
	// Verify updated_at was changed
	assert.True(suite.T(), updatedWallet.UpdatedAt.After(initialWallet.UpdatedAt))
}

func (suite *WalletRepositoryIntegrationTestSuite) TestSetTONWalletAddress_UpdateExisting() {
	ctx := context.Background()
	
	// Create wallet with existing TON address
	oldAddress := "EQD4FPq-PRDieyQKkizFTRtSDyucUIqrj0v_zXJmqaDp6_0t"
	initialWallet := &models.Wallet{
		UserID:           suite.testUserID,
		TonBalance:       decimal.NewFromFloat(100.00),
		FuelBalance:      decimal.NewFromFloat(200.00),
		BurnBalance:      decimal.NewFromFloat(0.00),
		TonWalletAddress: &oldAddress,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, initialWallet)
	require.NoError(suite.T(), err)
	
	// Add a small delay to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)
	
	// Update TON wallet address
	newAddress := "EQBvW8Z5huBkMJYdnfAEM5JqTNkuWX3diqYENkWsIL0XggGG"
	err = suite.walletRepo.SetTONWalletAddress(ctx, suite.testUserID, newAddress)
	require.NoError(suite.T(), err)
	
	// Verify address was updated
	updatedWallet, err := suite.walletRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), updatedWallet)
	
	require.NotNil(suite.T(), updatedWallet.TonWalletAddress)
	assert.Equal(suite.T(), newAddress, *updatedWallet.TonWalletAddress)
	assert.NotEqual(suite.T(), oldAddress, *updatedWallet.TonWalletAddress)
}

func (suite *WalletRepositoryIntegrationTestSuite) TestSetTONWalletAddress_NonExistentUser() {
	ctx := context.Background()
	
	nonExistentUserID := uuid.New()
	address := "EQD4FPq-PRDieyQKkizFTRtSDyucUIqrj0v_zXJmqaDp6_0t"
	
	err := suite.walletRepo.SetTONWalletAddress(ctx, nonExistentUserID, address)
	require.NoError(suite.T(), err) // UPDATE with no matching rows doesn't error
	
	// Verify no wallet was created
	wallet, err := suite.walletRepo.GetByUserID(ctx, nonExistentUserID)
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), wallet)
}

func (suite *WalletRepositoryIntegrationTestSuite) TestCreate_WithNullOptionalFields() {
	ctx := context.Background()
	
	wallet := &models.Wallet{
		UserID:               suite.testUserID,
		TonBalance:           decimal.NewFromFloat(0.00),
		FuelBalance:          decimal.NewFromFloat(1000.00), // Initial FUEL balance
		BurnBalance:          decimal.NewFromFloat(0.00),
		RookieRacesCompleted: 0,
		TonWalletAddress:     nil, // No TON wallet connected yet
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, wallet)
	require.NoError(suite.T(), err)
	
	// Verify wallet was created with null TON address
	retrievedWallet, err := suite.walletRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedWallet)
	
	assert.Nil(suite.T(), retrievedWallet.TonWalletAddress)
	assert.True(suite.T(), decimal.NewFromFloat(1000.00).Equal(retrievedWallet.FuelBalance))
}

func (suite *WalletRepositoryIntegrationTestSuite) TestUpdateBalances_ZeroDeltas() {
	ctx := context.Background()
	
	// Create initial wallet
	initialWallet := &models.Wallet{
		UserID:      suite.testUserID,
		TonBalance:  decimal.NewFromFloat(100.00),
		FuelBalance: decimal.NewFromFloat(200.00),
		BurnBalance: decimal.NewFromFloat(50.00),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	
	err := suite.walletRepo.Create(ctx, initialWallet)
	require.NoError(suite.T(), err)
	
	// Add a small delay to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)
	
	// Update with zero deltas
	zeroDelta := decimal.NewFromFloat(0.00)
	err = suite.walletRepo.UpdateBalances(ctx, suite.testUserID, zeroDelta, zeroDelta, zeroDelta)
	require.NoError(suite.T(), err)
	
	// Verify balances remain the same but updated_at changed
	updatedWallet, err := suite.walletRepo.GetByUserID(ctx, suite.testUserID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), updatedWallet)
	
	assert.True(suite.T(), initialWallet.TonBalance.Equal(updatedWallet.TonBalance))
	assert.True(suite.T(), initialWallet.FuelBalance.Equal(updatedWallet.FuelBalance))
	assert.True(suite.T(), initialWallet.BurnBalance.Equal(updatedWallet.BurnBalance))
	
	// Verify updated_at was still changed
	assert.True(suite.T(), updatedWallet.UpdatedAt.After(initialWallet.UpdatedAt))
}