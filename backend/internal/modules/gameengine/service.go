package gameengine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/storage/postgres/models"
	"github.com/megaherz/ndr/internal/storage/postgres/repository"
)

// GameEngineService handles game engine operations
type GameEngineService interface {
	// CreateMatch creates a new match with the given players
	CreateMatch(ctx context.Context, league string, players []*MatchPlayer) (*models.Match, error)

	// GetMatch retrieves a match by ID
	GetMatch(ctx context.Context, matchID uuid.UUID) (*models.Match, error)

	// StartMatch starts a match (transitions from FORMING to IN_PROGRESS)
	StartMatch(ctx context.Context, matchID uuid.UUID) error

	// EarnPoints locks a player's score for the current heat
	EarnPoints(ctx context.Context, matchID, userID uuid.UUID, score decimal.Decimal) error

	// GetMatchState retrieves the current match state
	GetMatchState(ctx context.Context, matchID uuid.UUID) (*MatchState, error)

	// CompleteMatch completes a match and triggers settlement
	CompleteMatch(ctx context.Context, matchID uuid.UUID) error
}

// MatchPlayer represents a player participating in a match
type MatchPlayer struct {
	UserID        *uuid.UUID      `json:"user_id,omitempty"` // Null for ghosts
	DisplayName   string          `json:"display_name"`
	IsGhost       bool            `json:"is_ghost"`
	GhostReplayID *uuid.UUID      `json:"ghost_replay_id,omitempty"`
	BuyinAmount   decimal.Decimal `json:"buyin_amount"`
}

// MatchState represents the current state of a match
type MatchState struct {
	MatchID       uuid.UUID      `json:"match_id"`
	League        string         `json:"league"`
	Status        string         `json:"status"`
	CurrentHeat   int            `json:"current_heat"` // 1, 2, or 3
	HeatStatus    HeatStatus     `json:"heat_status"`
	Players       []*PlayerState `json:"players"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	CrashSeed     string         `json:"crash_seed"`
	CrashSeedHash string         `json:"crash_seed_hash"`
}

// PlayerState represents a player's state in a match
type PlayerState struct {
	UserID      *uuid.UUID       `json:"user_id,omitempty"`
	DisplayName string           `json:"display_name"`
	IsGhost     bool             `json:"is_ghost"`
	Heat1Score  *decimal.Decimal `json:"heat1_score,omitempty"`
	Heat2Score  *decimal.Decimal `json:"heat2_score,omitempty"`
	Heat3Score  *decimal.Decimal `json:"heat3_score,omitempty"`
	TotalScore  decimal.Decimal  `json:"total_score"`
	Position    int              `json:"position"`
	IsAlive     bool             `json:"is_alive"`   // False if crashed in current heat
	HasLocked   bool             `json:"has_locked"` // True if locked score in current heat
}

// gameEngineService implements GameEngineService
type gameEngineService struct {
	matchRepo       repository.MatchRepository
	participantRepo repository.MatchParticipantRepository
	fairnessEngine  ProvableFairnessEngine
	physicsEngine   PhysicsEngine
	logger          *logrus.Logger
}

// NewGameEngineService creates a new game engine service
func NewGameEngineService(
	matchRepo repository.MatchRepository,
	participantRepo repository.MatchParticipantRepository,
	logger *logrus.Logger,
) GameEngineService {
	return &gameEngineService{
		matchRepo:       matchRepo,
		participantRepo: participantRepo,
		fairnessEngine:  NewProvableFairnessEngine(),
		physicsEngine:   NewPhysicsEngine(),
		logger:          logger,
	}
}

// CreateMatch creates a new match with the given players
func (s *gameEngineService) CreateMatch(ctx context.Context, league string, players []*MatchPlayer) (*models.Match, error) {
	if len(players) != 10 {
		return nil, fmt.Errorf("match must have exactly 10 players, got %d", len(players))
	}

	// Generate crash seeds for provable fairness
	matchID := uuid.New()
	seedData, commitHash, err := GenerateMatchSeeds(matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate crash seeds: %w", err)
	}

	// Serialize seed data
	seedJSON, err := json.Marshal(seedData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize seed data: %w", err)
	}

	// Calculate prize pool and rake
	totalBuyin := decimal.Zero
	livePlayerCount := 0
	ghostPlayerCount := 0

	for _, player := range players {
		totalBuyin = totalBuyin.Add(player.BuyinAmount)
		if player.IsGhost {
			ghostPlayerCount++
		} else {
			livePlayerCount++
		}
	}

	// 8% rake
	rakeAmount := totalBuyin.Mul(decimal.NewFromFloat(0.08)).Truncate(2)
	prizePool := totalBuyin.Sub(rakeAmount)

	// Create match
	match := &models.Match{
		ID:               matchID,
		League:           models.League(league),
		Status:           models.MatchStatusForming,
		LivePlayerCount:  livePlayerCount,
		GhostPlayerCount: ghostPlayerCount,
		PrizePool:        prizePool,
		RakeAmount:       rakeAmount,
		CrashSeed:        string(seedJSON),
		CrashSeedHash:    commitHash,
		StartedAt:        nil,
		CompletedAt:      nil,
		CreatedAt:        time.Now(),
	}

	// Save match to database
	err = s.matchRepo.Create(ctx, match)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"league":   league,
			"error":    err,
		}).Error("Failed to create match")
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	// Create match participants
	participants := make([]*models.MatchParticipant, 0, 10)
	for _, player := range players {
		participant := &models.MatchParticipant{
			MatchID:           matchID,
			UserID:            player.UserID,
			IsGhost:           player.IsGhost,
			GhostReplayID:     player.GhostReplayID,
			PlayerDisplayName: player.DisplayName,
			BuyinAmount:       player.BuyinAmount,
			Heat1Score:        nil,
			Heat2Score:        nil,
			Heat3Score:        nil,
			TotalScore:        nil,
			FinalPosition:     nil,
			PrizeAmount:       decimal.Zero,
			BurnReward:        decimal.Zero,
			CreatedAt:         time.Now(),
		}
		participants = append(participants, participant)
	}

	// Save participants to database
	err = s.participantRepo.CreateBatch(ctx, participants)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"error":    err,
		}).Error("Failed to create match participants")
		return nil, fmt.Errorf("failed to create match participants: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"match_id":      matchID,
		"league":        league,
		"live_players":  livePlayerCount,
		"ghost_players": ghostPlayerCount,
		"prize_pool":    prizePool,
		"rake_amount":   rakeAmount,
	}).Info("Match created successfully")

	return match, nil
}

// GetMatch retrieves a match by ID
func (s *gameEngineService) GetMatch(ctx context.Context, matchID uuid.UUID) (*models.Match, error) {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if match == nil {
		return nil, fmt.Errorf("match not found: %s", matchID)
	}

	return match, nil
}

// StartMatch starts a match (transitions from FORMING to IN_PROGRESS)
func (s *gameEngineService) StartMatch(ctx context.Context, matchID uuid.UUID) error {
	// Update match status
	err := s.matchRepo.UpdateStatus(ctx, matchID, string(models.MatchStatusInProgress))
	if err != nil {
		return fmt.Errorf("failed to update match status: %w", err)
	}

	// Set start time
	err = s.matchRepo.SetStartTime(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to set match start time: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"match_id": matchID,
	}).Info("Match started")

	// TODO: Start heat lifecycle (Heat 1)
	// This would be handled by the heat manager

	return nil
}

// EarnPoints locks a player's score for the current heat
func (s *gameEngineService) EarnPoints(ctx context.Context, matchID, userID uuid.UUID, score decimal.Decimal) error {
	// Get match to determine current heat
	match, err := s.GetMatch(ctx, matchID)
	if err != nil {
		return err
	}

	if match.Status != models.MatchStatusInProgress {
		return fmt.Errorf("match is not in progress")
	}

	// TODO: Determine current heat from match state
	// For now, assume Heat 1
	currentHeat := 1

	// Validate score is achievable (anti-cheat)
	if !s.physicsEngine.IsValidSpeed(score) {
		return fmt.Errorf("invalid score: %s", score.String())
	}

	// Update participant score
	err = s.participantRepo.UpdateHeatScore(ctx, matchID, userID, currentHeat, score)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"user_id":  userID,
			"heat":     currentHeat,
			"score":    score,
			"error":    err,
		}).Error("Failed to update heat score")
		return fmt.Errorf("failed to update heat score: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"user_id":  userID,
		"heat":     currentHeat,
		"score":    score,
	}).Info("Player earned points")

	// TODO: Check if all players have locked scores (early heat end)
	// TODO: Update total score if this was the final heat

	return nil
}

// GetMatchState retrieves the current match state
func (s *gameEngineService) GetMatchState(ctx context.Context, matchID uuid.UUID) (*MatchState, error) {
	// Get match
	match, err := s.GetMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	// Get participants
	participants, err := s.participantRepo.GetByMatchID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match participants: %w", err)
	}

	// Build player states
	playerStates := make([]*PlayerState, 0, len(participants))
	for _, participant := range participants {
		playerState := &PlayerState{
			UserID:      participant.UserID,
			DisplayName: participant.PlayerDisplayName,
			IsGhost:     participant.IsGhost,
			Heat1Score:  participant.Heat1Score,
			Heat2Score:  participant.Heat2Score,
			Heat3Score:  participant.Heat3Score,
			TotalScore: func() decimal.Decimal {
				if participant.TotalScore != nil {
					return *participant.TotalScore
				}
				return decimal.Zero
			}(),
			Position:  0,     // TODO: Calculate position
			IsAlive:   true,  // TODO: Determine from current heat state
			HasLocked: false, // TODO: Determine from current heat state
		}
		playerStates = append(playerStates, playerState)
	}

	// Build match state
	matchState := &MatchState{
		MatchID:       matchID,
		League:        string(match.League),
		Status:        string(match.Status),
		CurrentHeat:   1,                 // TODO: Determine from match state
		HeatStatus:    HeatStatusWaiting, // TODO: Determine from heat manager
		Players:       playerStates,
		StartedAt:     match.StartedAt,
		CompletedAt:   match.CompletedAt,
		CrashSeed:     match.CrashSeed,
		CrashSeedHash: match.CrashSeedHash,
	}

	return matchState, nil
}

// CompleteMatch completes a match and triggers settlement
func (s *gameEngineService) CompleteMatch(ctx context.Context, matchID uuid.UUID) error {
	// Update match status
	err := s.matchRepo.UpdateStatus(ctx, matchID, string(models.MatchStatusCompleted))
	if err != nil {
		return fmt.Errorf("failed to update match status: %w", err)
	}

	// Set completion time
	err = s.matchRepo.SetCompletionTime(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to set match completion time: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"match_id": matchID,
	}).Info("Match completed")

	// TODO: Trigger settlement process
	// This would be handled by the settlement module

	return nil
}
