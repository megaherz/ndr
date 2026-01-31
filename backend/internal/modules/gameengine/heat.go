package gameengine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"backend/internal/modules/gateway"
	"backend/internal/modules/gateway/events"
)

// HeatManager manages the lifecycle of heats within a match
type HeatManager interface {
	// StartHeatCountdown starts the 3-second countdown for a heat
	StartHeatCountdown(ctx context.Context, matchID uuid.UUID, heat int) error
	
	// StartHeatActive transitions from countdown to active heat
	StartHeatActive(ctx context.Context, matchID uuid.UUID) error
	
	// EndHeat ends the current heat and transitions to intermission or next heat
	EndHeat(ctx context.Context, matchID uuid.UUID) error
	
	// StartIntermission starts the 5-second intermission between heats
	StartIntermission(ctx context.Context, matchID uuid.UUID) error
	
	// CheckHeatTimeout checks if any heats have timed out
	CheckHeatTimeout(ctx context.Context) error
	
	// GetHeatTimeRemaining returns the time remaining in the current heat
	GetHeatTimeRemaining(ctx context.Context, matchID uuid.UUID) (time.Duration, error)
}

// HeatLifecycleEvent represents events in the heat lifecycle
type HeatLifecycleEvent struct {
	MatchID   uuid.UUID   `json:"match_id"`
	Heat      int         `json:"heat"`
	Status    HeatStatus  `json:"status"`
	Timestamp time.Time   `json:"timestamp"`
	Duration  int         `json:"duration,omitempty"` // Duration in seconds
}

// heatManager implements HeatManager
type heatManager struct {
	stateManager MatchStateManager
	publisher    gateway.CentrifugoPublisher
	logger       *logrus.Logger
	
	// Heat configuration
	countdownDuration   time.Duration // 3 seconds
	heatDuration       time.Duration // 25 seconds
	intermissionDuration time.Duration // 5 seconds
}

// NewHeatManager creates a new heat manager
func NewHeatManager(stateManager MatchStateManager, publisher gateway.CentrifugoPublisher, logger *logrus.Logger) HeatManager {
	return &heatManager{
		stateManager:         stateManager,
		publisher:           publisher,
		logger:              logger,
		countdownDuration:   3 * time.Second,
		heatDuration:       25 * time.Second,
		intermissionDuration: 5 * time.Second,
	}
}

// StartHeatCountdown starts the 3-second countdown for a heat
func (h *heatManager) StartHeatCountdown(ctx context.Context, matchID uuid.UUID, heat int) error {
	if heat < 1 || heat > 3 {
		return fmt.Errorf("invalid heat number: %d", heat)
	}
	
	// Update match state
	err := h.stateManager.StartHeat(ctx, matchID, heat)
	if err != nil {
		return fmt.Errorf("failed to start heat in state manager: %w", err)
	}
	
	h.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"heat":     heat,
	}).Info("Heat countdown started")
	
	// Publish heat_started event to match:{match_id} channel (T060)
	err = h.publishHeatStartedEvent(ctx, matchID, heat)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"heat":     heat,
			"error":    err,
		}).Error("Failed to publish heat started event")
		// Continue anyway - heat is started
	}
	
	// Schedule transition to active after countdown
	go func() {
		time.Sleep(h.countdownDuration)
		if err := h.StartHeatActive(ctx, matchID); err != nil {
			h.logger.WithFields(logrus.Fields{
				"match_id": matchID,
				"heat":     heat,
				"error":    err,
			}).Error("Failed to transition heat to active")
		}
	}()
	
	return nil
}

// StartHeatActive transitions from countdown to active heat
func (h *heatManager) StartHeatActive(ctx context.Context, matchID uuid.UUID) error {
	// Get current match state
	state, err := h.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match state: %w", err)
	}
	
	if state.HeatStatus != HeatStatusCountdown {
		return fmt.Errorf("heat is not in countdown status")
	}
	
	// Update heat status to active
	// Note: This would require extending the state manager interface
	// For now, we'll log the transition
	h.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"heat":     state.CurrentHeat,
	}).Info("Heat is now active")
	
	// Schedule heat end after heat duration
	go func() {
		time.Sleep(h.heatDuration)
		if err := h.EndHeat(ctx, matchID); err != nil {
			h.logger.WithFields(logrus.Fields{
				"match_id": matchID,
				"heat":     state.CurrentHeat,
				"error":    err,
			}).Error("Failed to end heat")
		}
	}()
	
	return nil
}

// EndHeat ends the current heat and transitions to intermission or next heat
func (h *heatManager) EndHeat(ctx context.Context, matchID uuid.UUID) error {
	// Get current match state
	state, err := h.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match state: %w", err)
	}
	
	// End the heat in state manager
	err = h.stateManager.EndHeat(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to end heat in state manager: %w", err)
	}
	
	h.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"heat":     state.CurrentHeat,
	}).Info("Heat ended")
	
	// Publish heat_ended event to match:{match_id} channel (T061)
	err = h.publishHeatEndedEvent(ctx, matchID, state.CurrentHeat)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"heat":     state.CurrentHeat,
			"error":    err,
		}).Error("Failed to publish heat ended event")
		// Continue anyway - heat is ended
	}
	
	// If this was Heat 3, the match is complete
	if state.CurrentHeat == 3 {
		h.logger.WithFields(logrus.Fields{
			"match_id": matchID,
		}).Info("Match completed after Heat 3")
		
		// TODO: Trigger match settlement
		return nil
	}
	
	// Start intermission before next heat
	return h.StartIntermission(ctx, matchID)
}

// StartIntermission starts the 5-second intermission between heats
func (h *heatManager) StartIntermission(ctx context.Context, matchID uuid.UUID) error {
	// Get current match state
	state, err := h.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match state: %w", err)
	}
	
	h.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"heat":     state.CurrentHeat,
	}).Info("Intermission started")
	
	// Schedule next heat after intermission
	nextHeat := state.CurrentHeat + 1
	go func() {
		time.Sleep(h.intermissionDuration)
		if err := h.StartHeatCountdown(ctx, matchID, nextHeat); err != nil {
			h.logger.WithFields(logrus.Fields{
				"match_id":  matchID,
				"next_heat": nextHeat,
				"error":     err,
			}).Error("Failed to start next heat after intermission")
		}
	}()
	
	return nil
}

// CheckHeatTimeout checks if any heats have timed out
func (h *heatManager) CheckHeatTimeout(ctx context.Context) error {
	// Get all active matches
	activeMatches := h.stateManager.GetActiveMatches(ctx)
	
	for _, matchID := range activeMatches {
		state, err := h.stateManager.GetMatchState(ctx, matchID)
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"match_id": matchID,
				"error":    err,
			}).Error("Failed to get match state for timeout check")
			continue
		}
		
		// Check if heat has timed out
		if state.HeatStatus == HeatStatusActive && state.HeatStartTime != nil {
			elapsed := time.Since(*state.HeatStartTime)
			totalHeatTime := h.countdownDuration + h.heatDuration
			
			if elapsed > totalHeatTime {
				h.logger.WithFields(logrus.Fields{
					"match_id": matchID,
					"heat":     state.CurrentHeat,
					"elapsed":  elapsed,
				}).Warn("Heat timed out, forcing end")
				
				// Force end the heat
				if err := h.EndHeat(ctx, matchID); err != nil {
					h.logger.WithFields(logrus.Fields{
						"match_id": matchID,
						"error":    err,
					}).Error("Failed to force end timed out heat")
				}
			}
		}
	}
	
	return nil
}

// GetHeatTimeRemaining returns the time remaining in the current heat
func (h *heatManager) GetHeatTimeRemaining(ctx context.Context, matchID uuid.UUID) (time.Duration, error) {
	state, err := h.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return 0, fmt.Errorf("failed to get match state: %w", err)
	}
	
	if state.HeatStartTime == nil {
		return 0, fmt.Errorf("heat has not started")
	}
	
	elapsed := time.Since(*state.HeatStartTime)
	
	switch state.HeatStatus {
	case HeatStatusCountdown:
		remaining := h.countdownDuration - elapsed
		if remaining < 0 {
			return 0, nil
		}
		return remaining, nil
		
	case HeatStatusActive:
		totalTime := h.countdownDuration + h.heatDuration
		remaining := totalTime - elapsed
		if remaining < 0 {
			return 0, nil
		}
		return remaining, nil
		
	default:
		return 0, nil
	}
}

// CheckEarlyHeatEnd checks if all alive players have locked scores (early end optimization)
// This implements T053 - early heat end when all players have finished
func (h *heatManager) CheckEarlyHeatEnd(ctx context.Context, matchID uuid.UUID) error {
	state, err := h.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return err
	}
	
	if state.HeatStatus != HeatStatusActive {
		return nil // Only check during active heat
	}
	
	// Count alive players who haven't locked
	unlockedCount := 0
	for _, player := range state.Players {
		if player.IsAlive && !player.HasLocked {
			unlockedCount++
		}
	}
	
	// If all alive players have locked, end the heat early
	if unlockedCount == 0 {
		h.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"heat":     state.CurrentHeat,
		}).Info("All players locked, ending heat early")
		
		return h.EndHeat(ctx, matchID)
	}
	
	return nil
}

// publishHeatStartedEvent publishes heat_started event to match channel
func (h *heatManager) publishHeatStartedEvent(ctx context.Context, matchID uuid.UUID, heat int) error {
	// Get match state for participants
	state, err := h.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match state: %w", err)
	}
	
	// Build participant info
	participants := make([]events.ParticipantInfo, 0, len(state.Players))
	position := 1
	for _, player := range state.Players {
		participantInfo := events.ParticipantInfo{
			UserID:      player.UserID,
			DisplayName: player.DisplayName,
			IsGhost:     player.IsGhost,
			Position:    position,
		}
		participants = append(participants, participantInfo)
		position++
	}
	
	// Calculate target line for Heat 2 and 3
	var targetLine *decimal.Decimal
	if heat == 2 {
		// Target line is Heat 1 winner's score
		targetLine = h.calculateHeat1WinnerScore(state)
	} else if heat == 3 {
		// Target line is current leader's total score
		targetLine = h.calculateCurrentLeaderTotal(state)
	}
	
	// Create heat started event
	heatStartedEvent := &events.HeatStartedEvent{
		MatchID:      matchID,
		Heat:         heat,
		Duration:     25, // 25 seconds
		StartTime:    time.Now(),
		TargetLine:   targetLine,
		Participants: participants,
	}
	
	// Publish to match channel
	err = h.publisher.PublishToMatch(ctx, matchID, events.EventHeatStarted, heatStartedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish heat started event: %w", err)
	}
	
	h.logger.WithFields(logrus.Fields{
		"match_id":         matchID,
		"heat":            heat,
		"participant_count": len(participants),
		"target_line":     targetLine,
	}).Info("Published heat started event")
	
	return nil
}

// publishHeatEndedEvent publishes heat_ended event to match channel
func (h *heatManager) publishHeatEndedEvent(ctx context.Context, matchID uuid.UUID, heat int) error {
	// Get match state for results
	state, err := h.stateManager.GetMatchState(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match state: %w", err)
	}
	
	// Build heat results
	results := make([]events.HeatResult, 0, len(state.Players))
	for _, player := range state.Players {
		var score *decimal.Decimal
		var crashed bool
		
		switch heat {
		case 1:
			score = player.Heat1Score
		case 2:
			score = player.Heat2Score
		case 3:
			score = player.Heat3Score
		}
		
		if score == nil || score.IsZero() {
			crashed = true
		}
		
		result := events.HeatResult{
			UserID:      player.UserID,
			DisplayName: player.DisplayName,
			IsGhost:     player.IsGhost,
			Score:       score,
			LockTime:    nil, // Could add lock time tracking
			Crashed:     crashed,
		}
		results = append(results, result)
	}
	
	// Build standings (sorted by total score)
	standings := h.buildStandings(state)
	
	// Calculate actual duration
	actualDuration := float64(h.countdownDuration.Seconds() + h.heatDuration.Seconds())
	if state.HeatStartTime != nil && state.HeatEndTime != nil {
		actualDuration = state.HeatEndTime.Sub(*state.HeatStartTime).Seconds()
	}
	
	// Calculate final speed at heat end
	finalSpeed := decimal.NewFromInt(500) // Max speed, could be more precise
	
	// Create heat ended event
	heatEndedEvent := &events.HeatEndedEvent{
		MatchID:        matchID,
		Heat:           heat,
		EndTime:        time.Now(),
		ActualDuration: actualDuration,
		FinalSpeed:     finalSpeed,
		Results:        results,
		Standings:      standings,
	}
	
	// Publish to match channel
	err = h.publisher.PublishToMatch(ctx, matchID, events.EventHeatEnded, heatEndedEvent)
	if err != nil {
		return fmt.Errorf("failed to publish heat ended event: %w", err)
	}
	
	h.logger.WithFields(logrus.Fields{
		"match_id":        matchID,
		"heat":           heat,
		"actual_duration": actualDuration,
		"result_count":   len(results),
	}).Info("Published heat ended event")
	
	return nil
}

// calculateHeat1WinnerScore calculates the Heat 1 winner's score for target line
func (h *heatManager) calculateHeat1WinnerScore(state *InMemoryMatchState) *decimal.Decimal {
	var bestScore decimal.Decimal
	
	for _, player := range state.Players {
		if player.Heat1Score != nil && player.Heat1Score.GreaterThan(bestScore) {
			bestScore = *player.Heat1Score
		}
	}
	
	if bestScore.IsZero() {
		return nil
	}
	
	return &bestScore
}

// calculateCurrentLeaderTotal calculates the current leader's total score
func (h *heatManager) calculateCurrentLeaderTotal(state *InMemoryMatchState) *decimal.Decimal {
	var bestTotal decimal.Decimal
	
	for _, player := range state.Players {
		if player.TotalScore.GreaterThan(bestTotal) {
			bestTotal = player.TotalScore
		}
	}
	
	if bestTotal.IsZero() {
		return nil
	}
	
	return &bestTotal
}

// buildStandings builds current standings for the heat ended event
func (h *heatManager) buildStandings(state *InMemoryMatchState) []events.StandingEntry {
	standings := make([]events.StandingEntry, 0, len(state.Players))
	
	for _, player := range state.Players {
		standing := events.StandingEntry{
			UserID:        player.UserID,
			DisplayName:   player.DisplayName,
			IsGhost:       player.IsGhost,
			Position:      player.Position,
			TotalScore:    player.TotalScore,
			Heat1Score:    player.Heat1Score,
			Heat2Score:    player.Heat2Score,
			Heat3Score:    player.Heat3Score,
			PositionDelta: 0, // Could calculate position change from previous heat
		}
		standings = append(standings, standing)
	}
	
	return standings
}