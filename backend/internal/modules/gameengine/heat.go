package gameengine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	logger       *logrus.Logger
	
	// Heat configuration
	countdownDuration   time.Duration // 3 seconds
	heatDuration       time.Duration // 25 seconds
	intermissionDuration time.Duration // 5 seconds
}

// NewHeatManager creates a new heat manager
func NewHeatManager(stateManager MatchStateManager, logger *logrus.Logger) HeatManager {
	return &heatManager{
		stateManager:         stateManager,
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
	
	// TODO: Publish heat_started event to match:{match_id} channel
	// This would be handled by the gateway publisher
	
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
	
	// TODO: Publish heat_ended event to match:{match_id} channel
	// This would include final standings and scores
	
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