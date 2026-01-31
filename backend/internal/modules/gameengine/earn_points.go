package gameengine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"backend/internal/storage/postgres/repository"
)

// EarnPointsService handles the logic for players locking their scores
type EarnPointsService interface {
	// LockScore locks a player's score for the current heat
	LockScore(ctx context.Context, matchID, userID uuid.UUID, requestedScore decimal.Decimal) (*EarnPointsResult, error)
	
	// GetCurrentHeatInfo returns information about the current heat
	GetCurrentHeatInfo(ctx context.Context, matchID uuid.UUID) (*HeatInfo, error)
	
	// ValidateScoreForTime validates that a score is achievable at the current time
	ValidateScoreForTime(ctx context.Context, matchID uuid.UUID, score decimal.Decimal) error
}

// EarnPointsResult represents the result of locking a score
type EarnPointsResult struct {
	Success      bool            `json:"success"`
	LockedScore  decimal.Decimal `json:"locked_score"`
	Heat         int             `json:"heat"`
	LockTime     time.Time       `json:"lock_time"`
	Position     int             `json:"position"`      // Position in current heat
	TotalScore   decimal.Decimal `json:"total_score"`   // Updated total score
	Message      string          `json:"message,omitempty"`
}

// HeatInfo represents information about the current heat
type HeatInfo struct {
	MatchID       uuid.UUID   `json:"match_id"`
	Heat          int         `json:"heat"`
	Status        HeatStatus  `json:"status"`
	StartTime     *time.Time  `json:"start_time,omitempty"`
	ElapsedTime   float64     `json:"elapsed_time"`   // Seconds since heat started
	MaxSpeed      decimal.Decimal `json:"max_speed"`  // Maximum achievable speed at current time
	PlayersLocked int         `json:"players_locked"` // Number of players who have locked
	TotalPlayers  int         `json:"total_players"`
}

// earnPointsService implements EarnPointsService
type earnPointsService struct {
	stateManager    MatchStateManager
	participantRepo repository.MatchParticipantRepository
	physicsEngine   PhysicsEngine
	heatManager     HeatManager
	logger          *logrus.Logger
}

// NewEarnPointsService creates a new earn points service
func NewEarnPointsService(
	stateManager MatchStateManager,
	participantRepo repository.MatchParticipantRepository,
	physicsEngine PhysicsEngine,
	heatManager HeatManager,
	logger *logrus.Logger,
) EarnPointsService {
	return &earnPointsService{
		stateManager:    stateManager,
		participantRepo: participantRepo,
		physicsEngine:   physicsEngine,
		heatManager:     heatManager,
		logger:          logger,
	}
}

// LockScore locks a player's score for the current heat
func (s *earnPointsService) LockScore(ctx context.Context, matchID, userID uuid.UUID, requestedScore decimal.Decimal) (*EarnPointsResult, error) {
	// Get match state
	state, err := s.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match state: %w", err)
	}
	
	// Validate match is in progress
	if state.Status != MatchStatusInProgress {
		return nil, fmt.Errorf("match is not in progress")
	}
	
	// Validate heat is active
	if state.HeatStatus != HeatStatusActive {
		return nil, fmt.Errorf("heat is not active (status: %s)", state.HeatStatus)
	}
	
	// Find player in match state
	var player *InMemoryPlayer
	for _, p := range state.Players {
		if p.UserID != nil && *p.UserID == userID {
			player = p
			break
		}
	}
	
	if player == nil {
		return nil, fmt.Errorf("player not found in match")
	}
	
	// Check if player already locked score for this heat
	if player.HasLocked {
		return nil, fmt.Errorf("player has already locked score for heat %d", state.CurrentHeat)
	}
	
	// Check if player is alive (hasn't crashed)
	if !player.IsAlive {
		return nil, fmt.Errorf("player has crashed and cannot lock score")
	}
	
	// Validate score against physics (anti-cheat)
	err = s.ValidateScoreForTime(ctx, matchID, requestedScore)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"user_id":  userID,
			"score":    requestedScore,
			"heat":     state.CurrentHeat,
			"error":    err,
		}).Warn("Invalid score attempt detected")
		return nil, fmt.Errorf("invalid score: %w", err)
	}
	
	// Lock the score in memory state
	lockTime := time.Now()
	err = s.stateManager.LockPlayerScore(ctx, matchID, userID, requestedScore)
	if err != nil {
		return nil, fmt.Errorf("failed to lock score in state: %w", err)
	}
	
	// Update score in database
	err = s.participantRepo.UpdateHeatScore(ctx, matchID, userID, state.CurrentHeat, requestedScore)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"user_id":  userID,
			"heat":     state.CurrentHeat,
			"score":    requestedScore,
			"error":    err,
		}).Error("Failed to update heat score in database")
		return nil, fmt.Errorf("failed to update score: %w", err)
	}
	
	// Calculate updated total score
	totalScore := s.calculatePlayerTotal(player, state.CurrentHeat, requestedScore)
	
	// Update total score in database
	err = s.participantRepo.UpdateTotalScore(ctx, matchID, userID, totalScore)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"match_id":    matchID,
			"user_id":     userID,
			"total_score": totalScore,
			"error":       err,
		}).Error("Failed to update total score in database")
	}
	
	// Calculate current position (simplified - could be more sophisticated)
	position := s.calculateCurrentPosition(state, userID, requestedScore)
	
	s.logger.WithFields(logrus.Fields{
		"match_id":    matchID,
		"user_id":     userID,
		"heat":        state.CurrentHeat,
		"score":       requestedScore,
		"total_score": totalScore,
		"position":    position,
	}).Info("Player locked score successfully")
	
	// Check if all players have locked (early heat end)
	go func() {
		if err := s.checkEarlyHeatEnd(ctx, matchID); err != nil {
			s.logger.WithFields(logrus.Fields{
				"match_id": matchID,
				"error":    err,
			}).Error("Failed to check early heat end")
		}
	}()
	
	return &EarnPointsResult{
		Success:     true,
		LockedScore: requestedScore,
		Heat:        state.CurrentHeat,
		LockTime:    lockTime,
		Position:    position,
		TotalScore:  totalScore,
		Message:     fmt.Sprintf("Score locked at %.2f for Heat %d", requestedScore, state.CurrentHeat),
	}, nil
}

// GetCurrentHeatInfo returns information about the current heat
func (s *earnPointsService) GetCurrentHeatInfo(ctx context.Context, matchID uuid.UUID) (*HeatInfo, error) {
	state, err := s.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match state: %w", err)
	}
	
	var elapsedTime float64
	var maxSpeed decimal.Decimal
	
	if state.HeatStartTime != nil {
		elapsedTime = time.Since(*state.HeatStartTime).Seconds()
		maxSpeed = s.physicsEngine.CalculateSpeed(elapsedTime)
	}
	
	// Count locked players
	playersLocked := 0
	totalPlayers := len(state.Players)
	for _, player := range state.Players {
		if player.HasLocked {
			playersLocked++
		}
	}
	
	return &HeatInfo{
		MatchID:       matchID,
		Heat:          state.CurrentHeat,
		Status:        state.HeatStatus,
		StartTime:     state.HeatStartTime,
		ElapsedTime:   elapsedTime,
		MaxSpeed:      maxSpeed,
		PlayersLocked: playersLocked,
		TotalPlayers:  totalPlayers,
	}, nil
}

// ValidateScoreForTime validates that a score is achievable at the current time
func (s *earnPointsService) ValidateScoreForTime(ctx context.Context, matchID uuid.UUID, score decimal.Decimal) error {
	state, err := s.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return err
	}
	
	if state.HeatStartTime == nil {
		return fmt.Errorf("heat has not started")
	}
	
	// Calculate elapsed time since heat started (including countdown)
	elapsedTime := time.Since(*state.HeatStartTime).Seconds()
	
	// Subtract countdown time to get actual heat time
	countdownDuration := 3.0 // 3 seconds countdown
	actualHeatTime := elapsedTime - countdownDuration
	
	if actualHeatTime < 0 {
		return fmt.Errorf("heat is still in countdown")
	}
	
	// Get maximum possible speed at this time
	maxSpeed := s.physicsEngine.CalculateSpeed(actualHeatTime)
	
	// Allow small tolerance for network latency (0.1 seconds worth of speed)
	toleranceTime := 0.1
	maxSpeedWithTolerance := s.physicsEngine.CalculateSpeed(actualHeatTime + toleranceTime)
	
	if score.GreaterThan(maxSpeedWithTolerance) {
		return fmt.Errorf("score %.2f exceeds maximum possible speed %.2f at time %.2fs", 
			score, maxSpeed, actualHeatTime)
	}
	
	// Validate score is positive
	if score.LessThan(decimal.Zero) {
		return fmt.Errorf("score cannot be negative")
	}
	
	return nil
}

// calculatePlayerTotal calculates a player's total score across all heats
func (s *earnPointsService) calculatePlayerTotal(player *InMemoryPlayer, currentHeat int, newScore decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	
	// Add scores from completed heats
	if player.Heat1Score != nil {
		total = total.Add(*player.Heat1Score)
	}
	if player.Heat2Score != nil {
		total = total.Add(*player.Heat2Score)
	}
	if player.Heat3Score != nil {
		total = total.Add(*player.Heat3Score)
	}
	
	// Add the new score for current heat
	switch currentHeat {
	case 1:
		if player.Heat1Score == nil {
			total = total.Add(newScore)
		}
	case 2:
		if player.Heat2Score == nil {
			total = total.Add(newScore)
		}
	case 3:
		if player.Heat3Score == nil {
			total = total.Add(newScore)
		}
	}
	
	return total
}

// calculateCurrentPosition calculates a player's current position in the heat
func (s *earnPointsService) calculateCurrentPosition(state *InMemoryMatchState, userID uuid.UUID, score decimal.Decimal) int {
	// Count players with higher scores in current heat
	position := 1
	
	for _, player := range state.Players {
		if player.UserID != nil && *player.UserID == userID {
			continue // Skip self
		}
		
		var playerScore decimal.Decimal
		switch state.CurrentHeat {
		case 1:
			if player.Heat1Score != nil {
				playerScore = *player.Heat1Score
			}
		case 2:
			if player.Heat2Score != nil {
				playerScore = *player.Heat2Score
			}
		case 3:
			if player.Heat3Score != nil {
				playerScore = *player.Heat3Score
			}
		}
		
		if playerScore.GreaterThan(score) {
			position++
		}
	}
	
	return position
}

// checkEarlyHeatEnd checks if all alive players have locked scores
func (s *earnPointsService) checkEarlyHeatEnd(ctx context.Context, matchID uuid.UUID) error {
	state, err := s.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return err
	}
	
	if state.HeatStatus != HeatStatusActive {
		return nil // Only check during active heat
	}
	
	// Count alive players who haven't locked
	aliveUnlocked := 0
	for _, player := range state.Players {
		if player.IsAlive && !player.HasLocked {
			aliveUnlocked++
		}
	}
	
	// If all alive players have locked, trigger early heat end
	if aliveUnlocked == 0 {
		s.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"heat":     state.CurrentHeat,
		}).Info("All alive players locked, triggering early heat end")
		
		// Use heat manager to end the heat early
		return s.heatManager.EndHeat(ctx, matchID)
	}
	
	return nil
}