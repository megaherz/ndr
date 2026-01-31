package gameengine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/constants"
	"github.com/megaherz/ndr/internal/modules/account"
	"github.com/megaherz/ndr/internal/modules/gateway"
	"github.com/megaherz/ndr/internal/modules/gateway/events"
	"github.com/megaherz/ndr/internal/storage/postgres/models"
	"github.com/megaherz/ndr/internal/storage/postgres/repository"
)

// SettlementService handles match settlement and prize distribution
type SettlementService interface {
	// SettleMatch calculates final positions, distributes prizes, and applies ledger entries
	SettleMatch(ctx context.Context, matchID uuid.UUID) (*MatchSettlement, error)

	// CalculatePositions calculates final positions with tiebreaker logic
	CalculatePositions(ctx context.Context, matchID uuid.UUID) ([]*PlayerPosition, error)

	// CalculatePrizes calculates prize distribution based on positions
	CalculatePrizes(ctx context.Context, matchID uuid.UUID, positions []*PlayerPosition) (*PrizeDistribution, error)

	// ApplySettlement applies all ledger entries for the settlement
	ApplySettlement(ctx context.Context, matchID uuid.UUID, settlement *MatchSettlement) error
}

// MatchSettlement represents the complete settlement of a match
type MatchSettlement struct {
	MatchID           uuid.UUID             `json:"match_id"`
	League            string                `json:"league"`
	SettledAt         time.Time             `json:"settled_at"`
	Positions         []*PlayerPosition     `json:"positions"`
	PrizePool         decimal.Decimal       `json:"prize_pool"`
	RakeAmount        decimal.Decimal       `json:"rake_amount"`
	PrizeDistribution *PrizeDistribution    `json:"prize_distribution"`
	LedgerEntries     []*models.LedgerEntry `json:"ledger_entries"`
}

// PlayerPosition represents a player's final position and scores
type PlayerPosition struct {
	UserID        *uuid.UUID      `json:"user_id,omitempty"`
	DisplayName   string          `json:"display_name"`
	IsGhost       bool            `json:"is_ghost"`
	FinalPosition int             `json:"final_position"`
	Heat1Score    decimal.Decimal `json:"heat1_score"`
	Heat2Score    decimal.Decimal `json:"heat2_score"`
	Heat3Score    decimal.Decimal `json:"heat3_score"`
	TotalScore    decimal.Decimal `json:"total_score"`
	PrizeAmount   decimal.Decimal `json:"prize_amount"`
	BurnReward    decimal.Decimal `json:"burn_reward"`
}

// PrizeDistribution represents how prizes are distributed
type PrizeDistribution struct {
	TotalPrizePool decimal.Decimal         `json:"total_prize_pool"`
	FirstPlace     decimal.Decimal         `json:"first_place"`  // 50% of prize pool
	SecondPlace    decimal.Decimal         `json:"second_place"` // 30% of prize pool
	ThirdPlace     decimal.Decimal         `json:"third_place"`  // 20% of prize pool
	BurnRewards    map[int]decimal.Decimal `json:"burn_rewards"` // BURN rewards by position
}

// League-specific BURN reward tables
var burnRewardTables = map[string]map[int]decimal.Decimal{
	constants.LeagueRookie: {
		// Rookie league has no BURN rewards (FR-038)
	},
	constants.LeagueStreet: {
		1: decimal.NewFromInt(50), // 1st place: 50 BURN
		2: decimal.NewFromInt(30), // 2nd place: 30 BURN
		3: decimal.NewFromInt(20), // 3rd place: 20 BURN
		4: decimal.NewFromInt(10), // 4th place: 10 BURN
		5: decimal.NewFromInt(5),  // 5th place: 5 BURN
	},
	constants.LeaguePro: {
		1: decimal.NewFromInt(300), // 1st place: 300 BURN
		2: decimal.NewFromInt(200), // 2nd place: 200 BURN
		3: decimal.NewFromInt(150), // 3rd place: 150 BURN
		4: decimal.NewFromInt(100), // 4th place: 100 BURN
		5: decimal.NewFromInt(75),  // 5th place: 75 BURN
		6: decimal.NewFromInt(50),  // 6th place: 50 BURN
		7: decimal.NewFromInt(25),  // 7th place: 25 BURN
	},
	constants.LeagueTopFuel: {
		1:  decimal.NewFromInt(3000), // 1st place: 3000 BURN
		2:  decimal.NewFromInt(2000), // 2nd place: 2000 BURN
		3:  decimal.NewFromInt(1500), // 3rd place: 1500 BURN
		4:  decimal.NewFromInt(1000), // 4th place: 1000 BURN
		5:  decimal.NewFromInt(750),  // 5th place: 750 BURN
		6:  decimal.NewFromInt(500),  // 6th place: 500 BURN
		7:  decimal.NewFromInt(400),  // 7th place: 400 BURN
		8:  decimal.NewFromInt(300),  // 8th place: 300 BURN
		9:  decimal.NewFromInt(200),  // 9th place: 200 BURN
		10: decimal.NewFromInt(100),  // 10th place: 100 BURN
	},
}

// settlementService implements SettlementService
type settlementService struct {
	matchRepo       repository.MatchRepository
	participantRepo repository.MatchParticipantRepository
	settlementRepo  repository.MatchSettlementRepository
	ledgerOps       account.LedgerOperations
	stateManager    MatchStateManager
	publisher       gateway.CentrifugoPublisher
	logger          *logrus.Logger
}

// NewSettlementService creates a new settlement service
func NewSettlementService(
	matchRepo repository.MatchRepository,
	participantRepo repository.MatchParticipantRepository,
	settlementRepo repository.MatchSettlementRepository,
	ledgerOps account.LedgerOperations,
	stateManager MatchStateManager,
	publisher gateway.CentrifugoPublisher,
	logger *logrus.Logger,
) SettlementService {
	return &settlementService{
		matchRepo:       matchRepo,
		participantRepo: participantRepo,
		settlementRepo:  settlementRepo,
		ledgerOps:       ledgerOps,
		stateManager:    stateManager,
		publisher:       publisher,
		logger:          logger,
	}
}

// SettleMatch calculates final positions, distributes prizes, and applies ledger entries
func (s *settlementService) SettleMatch(ctx context.Context, matchID uuid.UUID) (*MatchSettlement, error) {
	// Get match information
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if match == nil {
		return nil, fmt.Errorf("match not found: %s", matchID)
	}

	// Check if already settled
	// This would require implementing MatchSettlementRepository first
	// For now, we'll skip this check

	// Calculate final positions
	positions, err := s.CalculatePositions(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate positions: %w", err)
	}

	// Calculate prizes
	prizeDistribution, err := s.CalculatePrizes(ctx, matchID, positions)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate prizes: %w", err)
	}

	// Apply prize amounts to positions
	s.applyPrizesToPositions(positions, prizeDistribution, string(match.League))

	// Update participant records with final positions and prizes
	err = s.updateParticipantResults(ctx, matchID, positions)
	if err != nil {
		return nil, fmt.Errorf("failed to update participant results: %w", err)
	}

	// Create settlement record
	settlement := &MatchSettlement{
		MatchID:           matchID,
		League:            string(match.League),
		SettledAt:         time.Now(),
		Positions:         positions,
		PrizePool:         match.PrizePool,
		RakeAmount:        match.RakeAmount,
		PrizeDistribution: prizeDistribution,
	}

	// Apply settlement to ledger
	err = s.ApplySettlement(ctx, matchID, settlement)
	if err != nil {
		return nil, fmt.Errorf("failed to apply settlement: %w", err)
	}

	// Update match status to completed
	err = s.matchRepo.UpdateStatus(ctx, matchID, string(models.MatchStatusCompleted))
	if err != nil {
		return nil, fmt.Errorf("failed to update match status: %w", err)
	}

	err = s.matchRepo.SetCompletionTime(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to set completion time: %w", err)
	}

	// Publish match settled event (T062)
	err = s.publishMatchSettledEvent(ctx, settlement)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"error":    err,
		}).Error("Failed to publish match settled event")
		// Continue anyway - settlement is complete
	}

	// Publish balance updated events to all live players (T063)
	err = s.publishBalanceUpdatedEvents(ctx, settlement)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"error":    err,
		}).Error("Failed to publish balance updated events")
		// Continue anyway - settlement is complete
	}

	s.logger.WithFields(logrus.Fields{
		"match_id":    matchID,
		"league":      match.League,
		"prize_pool":  match.PrizePool,
		"rake_amount": match.RakeAmount,
		"winner":      positions[0].DisplayName,
	}).Info("Match settled successfully")

	return settlement, nil
}

// CalculatePositions calculates final positions with tiebreaker logic
func (s *settlementService) CalculatePositions(ctx context.Context, matchID uuid.UUID) ([]*PlayerPosition, error) {
	// Get all participants
	participants, err := s.participantRepo.GetByMatchID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	// Convert to PlayerPosition structs
	positions := make([]*PlayerPosition, 0, len(participants))
	for _, p := range participants {
		position := &PlayerPosition{
			UserID:      p.UserID,
			DisplayName: p.PlayerDisplayName,
			IsGhost:     p.IsGhost,
			Heat1Score: func() decimal.Decimal {
				if p.Heat1Score != nil {
					return *p.Heat1Score
				}
				return decimal.Zero
			}(),
			Heat2Score: func() decimal.Decimal {
				if p.Heat2Score != nil {
					return *p.Heat2Score
				}
				return decimal.Zero
			}(),
			Heat3Score: func() decimal.Decimal {
				if p.Heat3Score != nil {
					return *p.Heat3Score
				}
				return decimal.Zero
			}(),
			TotalScore: func() decimal.Decimal {
				if p.TotalScore != nil {
					return *p.TotalScore
				}
				return decimal.Zero
			}(),
		}
		positions = append(positions, position)
	}

	// Sort positions using tiebreaker logic
	s.sortPositionsWithTiebreaker(positions)

	// Assign final positions
	for i, position := range positions {
		position.FinalPosition = i + 1
	}

	return positions, nil
}

// CalculatePrizes calculates prize distribution based on positions
func (s *settlementService) CalculatePrizes(ctx context.Context, matchID uuid.UUID, positions []*PlayerPosition) (*PrizeDistribution, error) {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	prizePool := match.PrizePool

	// Calculate FUEL prizes (top 3 only)
	firstPlace := prizePool.Mul(decimal.NewFromFloat(0.5)).Truncate(2)  // 50%
	secondPlace := prizePool.Mul(decimal.NewFromFloat(0.3)).Truncate(2) // 30%
	thirdPlace := prizePool.Mul(decimal.NewFromFloat(0.2)).Truncate(2)  // 20%

	// Get BURN rewards for this league
	burnRewards := burnRewardTables[string(match.League)]
	if burnRewards == nil {
		burnRewards = make(map[int]decimal.Decimal)
	}

	return &PrizeDistribution{
		TotalPrizePool: prizePool,
		FirstPlace:     firstPlace,
		SecondPlace:    secondPlace,
		ThirdPlace:     thirdPlace,
		BurnRewards:    burnRewards,
	}, nil
}

// ApplySettlement applies all ledger entries for the settlement
func (s *settlementService) ApplySettlement(ctx context.Context, matchID uuid.UUID, settlement *MatchSettlement) error {
	var ledgerEntries []*models.LedgerEntry

	// Create prize entries (FUEL)
	for _, position := range settlement.Positions {
		if position.UserID != nil && position.PrizeAmount.GreaterThan(decimal.Zero) {
			entry := &models.LedgerEntry{
				UserID:        position.UserID,
				SystemWallet:  nil,
				Currency:      constants.CurrencyFUEL,
				Amount:        position.PrizeAmount,
				OperationType: constants.OperationMatchPrize,
				ReferenceID:   &matchID,
				Description: func() *string {
					desc := fmt.Sprintf("Prize for position %d in %s league", position.FinalPosition, settlement.League)
					return &desc
				}(),
				CreatedAt: settlement.SettledAt,
			}
			ledgerEntries = append(ledgerEntries, entry)
		}

		// Create BURN reward entries
		if position.UserID != nil && position.BurnReward.GreaterThan(decimal.Zero) {
			entry := &models.LedgerEntry{
				UserID:        position.UserID,
				SystemWallet:  nil,
				Currency:      constants.CurrencyBURN,
				Amount:        position.BurnReward,
				OperationType: constants.OperationMatchBurnReward,
				ReferenceID:   &matchID,
				Description: func() *string {
					desc := fmt.Sprintf("BURN reward for position %d in %s league", position.FinalPosition, settlement.League)
					return &desc
				}(),
				CreatedAt: settlement.SettledAt,
			}
			ledgerEntries = append(ledgerEntries, entry)
		}
	}

	// Create rake entry (to RAKE_FUEL system wallet)
	if settlement.RakeAmount.GreaterThan(decimal.Zero) {
		entry := &models.LedgerEntry{
			UserID: nil,
			SystemWallet: func() *string {
				wallet := constants.SystemWalletRakeFuel
				return &wallet
			}(),
			Currency:      constants.CurrencyFUEL,
			Amount:        settlement.RakeAmount,
			OperationType: constants.OperationMatchRake,
			ReferenceID:   &matchID,
			Description: func() *string {
				desc := fmt.Sprintf("8%% rake from %s league match", settlement.League)
				return &desc
			}(),
			CreatedAt: settlement.SettledAt,
		}
		ledgerEntries = append(ledgerEntries, entry)
	}

	// Handle Ghost prize/buyin entries (to/from HOUSE_FUEL)
	for _, position := range settlement.Positions {
		if position.IsGhost {
			// Ghost won prize - debit from HOUSE_FUEL
			if position.PrizeAmount.GreaterThan(decimal.Zero) {
				entry := &models.LedgerEntry{
					UserID: nil,
					SystemWallet: func() *string {
						wallet := constants.SystemWalletHouseFuel
						return &wallet
					}(),
					Currency:      constants.CurrencyFUEL,
					Amount:        position.PrizeAmount.Neg(),
					OperationType: constants.OperationMatchPrize,
					ReferenceID:   &matchID,
					Description: func() *string {
						desc := fmt.Sprintf("Ghost prize payout for position %d", position.FinalPosition)
						return &desc
					}(),
					CreatedAt: settlement.SettledAt,
				}
				ledgerEntries = append(ledgerEntries, entry)
			}
		}
	}

	// Apply all ledger entries atomically
	err := s.ledgerOps.RecordMatchEntries(ctx, ledgerEntries)
	if err != nil {
		return fmt.Errorf("failed to record settlement ledger entries: %w", err)
	}

	settlement.LedgerEntries = ledgerEntries

	s.logger.WithFields(logrus.Fields{
		"match_id":    matchID,
		"entry_count": len(ledgerEntries),
		"prize_pool":  settlement.PrizePool,
		"rake_amount": settlement.RakeAmount,
	}).Info("Settlement ledger entries applied")

	return nil
}

// sortPositionsWithTiebreaker sorts positions using the tiebreaker logic
// Tiebreaker: Heat 3 score → Heat 2 score → Heat 1 score
func (s *settlementService) sortPositionsWithTiebreaker(positions []*PlayerPosition) {
	// Bubble sort with tiebreaker logic
	for i := 0; i < len(positions)-1; i++ {
		for j := i + 1; j < len(positions); j++ {
			if s.shouldSwapPositions(positions[i], positions[j]) {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}
}

// shouldSwapPositions determines if two positions should be swapped in sorting
func (s *settlementService) shouldSwapPositions(p1, p2 *PlayerPosition) bool {
	// First, compare total scores
	if p1.TotalScore.GreaterThan(p2.TotalScore) {
		return false // p1 is better
	}
	if p1.TotalScore.LessThan(p2.TotalScore) {
		return true // p2 is better
	}

	// Total scores are equal, use tiebreaker
	// Heat 3 tiebreaker
	if p1.Heat3Score.GreaterThan(p2.Heat3Score) {
		return false // p1 is better
	}
	if p1.Heat3Score.LessThan(p2.Heat3Score) {
		return true // p2 is better
	}

	// Heat 2 tiebreaker
	if p1.Heat2Score.GreaterThan(p2.Heat2Score) {
		return false // p1 is better
	}
	if p1.Heat2Score.LessThan(p2.Heat2Score) {
		return true // p2 is better
	}

	// Heat 1 tiebreaker
	return p1.Heat1Score.LessThan(p2.Heat1Score) // p2 is better if their Heat 1 score is higher
}

// applyPrizesToPositions applies prize amounts and BURN rewards to positions
func (s *settlementService) applyPrizesToPositions(positions []*PlayerPosition, prizes *PrizeDistribution, league string) {
	for _, position := range positions {
		// Apply FUEL prizes (top 3 only)
		switch position.FinalPosition {
		case 1:
			position.PrizeAmount = prizes.FirstPlace
		case 2:
			position.PrizeAmount = prizes.SecondPlace
		case 3:
			position.PrizeAmount = prizes.ThirdPlace
		default:
			position.PrizeAmount = decimal.Zero
		}

		// Apply BURN rewards (if not ghost and league has rewards)
		if !position.IsGhost {
			if burnAmount, exists := prizes.BurnRewards[position.FinalPosition]; exists {
				position.BurnReward = burnAmount
			} else {
				position.BurnReward = decimal.Zero
			}
		}
	}
}

// updateParticipantResults updates participant records with final results
func (s *settlementService) updateParticipantResults(ctx context.Context, matchID uuid.UUID, positions []*PlayerPosition) error {
	for _, position := range positions {
		if position.UserID == nil {
			continue // Skip ghosts (they don't have user IDs to update)
		}

		// Update final position
		err := s.participantRepo.SetFinalPosition(ctx, matchID, *position.UserID, position.FinalPosition)
		if err != nil {
			return fmt.Errorf("failed to set final position for user %s: %w", position.UserID, err)
		}

		// Update prize amount
		err = s.participantRepo.SetPrizeAmount(ctx, matchID, *position.UserID, position.PrizeAmount)
		if err != nil {
			return fmt.Errorf("failed to set prize amount for user %s: %w", position.UserID, err)
		}

		// Update BURN reward
		err = s.participantRepo.SetBurnReward(ctx, matchID, *position.UserID, position.BurnReward)
		if err != nil {
			return fmt.Errorf("failed to set burn reward for user %s: %w", position.UserID, err)
		}
	}

	return nil
}

// publishMatchSettledEvent publishes match_settled event to match channel (T062)
func (s *settlementService) publishMatchSettledEvent(ctx context.Context, settlement *MatchSettlement) error {
	// Build final standings
	finalStandings := make([]events.FinalStanding, 0, len(settlement.Positions))
	for _, position := range settlement.Positions {
		standing := events.FinalStanding{
			UserID:        position.UserID,
			DisplayName:   position.DisplayName,
			IsGhost:       position.IsGhost,
			FinalPosition: position.FinalPosition,
			TotalScore:    position.TotalScore,
			Heat1Score:    position.Heat1Score,
			Heat2Score:    position.Heat2Score,
			Heat3Score:    position.Heat3Score,
			PrizeAmount:   position.PrizeAmount,
			BurnReward:    position.BurnReward,
		}
		finalStandings = append(finalStandings, standing)
	}

	// Build prize distribution
	prizeDistribution := make([]events.PrizeEntry, 0)
	for _, position := range settlement.Positions {
		if position.PrizeAmount.GreaterThan(decimal.Zero) || position.BurnReward.GreaterThan(decimal.Zero) {
			entry := events.PrizeEntry{
				Position:    position.FinalPosition,
				UserID:      position.UserID,
				DisplayName: position.DisplayName,
				IsGhost:     position.IsGhost,
				PrizeAmount: position.PrizeAmount,
				BurnReward:  position.BurnReward,
			}
			prizeDistribution = append(prizeDistribution, entry)
		}
	}

	// Create match settled event
	matchSettledEvent := &events.MatchSettledEvent{
		MatchID:           settlement.MatchID,
		League:            settlement.League,
		CompletedAt:       settlement.SettledAt,
		FinalStandings:    finalStandings,
		PrizeDistribution: prizeDistribution,
		CrashSeed:         "revealed_seed_data", // TODO: Get actual crash seed from match
		CrashSeedHash:     "original_hash",      // TODO: Get actual hash from match
	}

	// Publish to match channel
	err := s.publisher.PublishToMatch(ctx, settlement.MatchID, events.EventMatchSettled, matchSettledEvent)
	if err != nil {
		return fmt.Errorf("failed to publish match settled event: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"match_id":      settlement.MatchID,
		"league":        settlement.League,
		"winner":        finalStandings[0].DisplayName,
		"prize_entries": len(prizeDistribution),
	}).Info("Published match settled event")

	return nil
}

// publishBalanceUpdatedEvents publishes balance_updated events to all live players (T063)
func (s *settlementService) publishBalanceUpdatedEvents(ctx context.Context, settlement *MatchSettlement) error {
	for _, position := range settlement.Positions {
		// Only publish to live players (not ghosts)
		if position.UserID == nil || position.IsGhost {
			continue
		}

		// Calculate balance changes
		changes := events.BalanceChanges{
			TONDelta:  decimal.Zero,         // No TON changes from matches
			FuelDelta: position.PrizeAmount, // FUEL prize (could be zero)
			BurnDelta: position.BurnReward,  // BURN reward (could be zero)
		}

		// Skip if no changes
		if changes.FuelDelta.IsZero() && changes.BurnDelta.IsZero() {
			continue
		}

		// Get updated balances (would need to query wallet)
		// For now, we'll use placeholder values
		balanceUpdatedEvent := &events.BalanceUpdatedEvent{
			UserID:      *position.UserID,
			TONBalance:  decimal.Zero, // TODO: Get actual balance
			FuelBalance: decimal.Zero, // TODO: Get actual balance
			BurnBalance: decimal.Zero, // TODO: Get actual balance
			Changes:     changes,
			Reason:      "match_settlement",
			ReferenceID: &settlement.MatchID,
		}

		// Publish to user's personal channel
		err := s.publisher.PublishToUser(ctx, *position.UserID, events.EventBalanceUpdated, balanceUpdatedEvent)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"match_id": settlement.MatchID,
				"user_id":  *position.UserID,
				"error":    err,
			}).Error("Failed to publish balance updated event to player")
			// Continue with other players
		}
	}

	s.logger.WithFields(logrus.Fields{
		"match_id": settlement.MatchID,
		"league":   settlement.League,
	}).Info("Published balance updated events to all live players")

	return nil
}
