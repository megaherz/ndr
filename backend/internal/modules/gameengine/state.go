package gameengine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// MatchStatus represents the status of a match
type MatchStatus string

const (
	MatchStatusForming     MatchStatus = "FORMING"
	MatchStatusInProgress  MatchStatus = "IN_PROGRESS"
	MatchStatusCompleted   MatchStatus = "COMPLETED"
	MatchStatusAborted     MatchStatus = "ABORTED"
)

// HeatStatus represents the status of a heat
type HeatStatus string

const (
	HeatStatusWaiting      HeatStatus = "WAITING"      // Before heat starts
	HeatStatusCountdown    HeatStatus = "COUNTDOWN"    // 3-second countdown
	HeatStatusActive       HeatStatus = "ACTIVE"       // Heat in progress
	HeatStatusIntermission HeatStatus = "INTERMISSION" // Between heats
	HeatStatusCompleted    HeatStatus = "COMPLETED"    // Heat finished
)

// MatchStateManager manages in-memory match states
type MatchStateManager interface {
	// CreateMatchState creates a new match state
	CreateMatchState(ctx context.Context, matchID uuid.UUID, league string, players []*MatchPlayer) error
	
	// GetMatchState retrieves a match state
	GetMatchState(ctx context.Context, matchID uuid.UUID) (*InMemoryMatchState, error)
	
	// UpdateMatchStatus updates the match status
	UpdateMatchStatus(ctx context.Context, matchID uuid.UUID, status MatchStatus) error
	
	// StartHeat starts a specific heat
	StartHeat(ctx context.Context, matchID uuid.UUID, heat int) error
	
	// EndHeat ends the current heat
	EndHeat(ctx context.Context, matchID uuid.UUID) error
	
	// LockPlayerScore locks a player's score for the current heat
	LockPlayerScore(ctx context.Context, matchID, userID uuid.UUID, score decimal.Decimal) error
	
	// GetActiveMatches returns all active match IDs
	GetActiveMatches(ctx context.Context) []uuid.UUID
	
	// RemoveMatchState removes a match state from memory
	RemoveMatchState(ctx context.Context, matchID uuid.UUID) error
}

// InMemoryMatchState represents the in-memory state of a match
type InMemoryMatchState struct {
	MatchID       uuid.UUID                    `json:"match_id"`
	League        string                       `json:"league"`
	Status        MatchStatus                  `json:"status"`
	CurrentHeat   int                          `json:"current_heat"`    // 1, 2, or 3
	HeatStatus    HeatStatus                   `json:"heat_status"`
	HeatStartTime *time.Time                   `json:"heat_start_time,omitempty"`
	HeatEndTime   *time.Time                   `json:"heat_end_time,omitempty"`
	Players       map[uuid.UUID]*InMemoryPlayer `json:"players"`
	CreatedAt     time.Time                    `json:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at"`
	
	// Synchronization
	mu sync.RWMutex `json:"-"`
}

// InMemoryPlayer represents a player's state in memory
type InMemoryPlayer struct {
	UserID       *uuid.UUID       `json:"user_id,omitempty"`
	DisplayName  string           `json:"display_name"`
	IsGhost      bool             `json:"is_ghost"`
	GhostReplayID *uuid.UUID      `json:"ghost_replay_id,omitempty"`
	Heat1Score   *decimal.Decimal `json:"heat1_score,omitempty"`
	Heat2Score   *decimal.Decimal `json:"heat2_score,omitempty"`
	Heat3Score   *decimal.Decimal `json:"heat3_score,omitempty"`
	TotalScore   decimal.Decimal  `json:"total_score"`
	Position     int              `json:"position"`
	IsAlive      bool             `json:"is_alive"`     // False if crashed in current heat
	HasLocked    bool             `json:"has_locked"`   // True if locked score in current heat
	LockTime     *time.Time       `json:"lock_time,omitempty"` // When they locked
}

// matchStateManager implements MatchStateManager
type matchStateManager struct {
	states map[uuid.UUID]*InMemoryMatchState
	mu     sync.RWMutex
	logger *logrus.Logger
}

// NewMatchStateManager creates a new match state manager
func NewMatchStateManager(logger *logrus.Logger) MatchStateManager {
	return &matchStateManager{
		states: make(map[uuid.UUID]*InMemoryMatchState),
		logger: logger,
	}
}

// CreateMatchState creates a new match state
func (m *matchStateManager) CreateMatchState(ctx context.Context, matchID uuid.UUID, league string, players []*MatchPlayer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if state already exists
	if _, exists := m.states[matchID]; exists {
		return fmt.Errorf("match state already exists for match %s", matchID)
	}
	
	// Create player states
	playerStates := make(map[uuid.UUID]*InMemoryPlayer)
	for _, player := range players {
		var playerID uuid.UUID
		if player.UserID != nil {
			playerID = *player.UserID
		} else {
			// Generate a temporary ID for ghosts
			playerID = uuid.New()
		}
		
		playerState := &InMemoryPlayer{
			UserID:        player.UserID,
			DisplayName:   player.DisplayName,
			IsGhost:       player.IsGhost,
			GhostReplayID: player.GhostReplayID,
			Heat1Score:    nil,
			Heat2Score:    nil,
			Heat3Score:    nil,
			TotalScore:    decimal.Zero,
			Position:      0,
			IsAlive:       true,
			HasLocked:     false,
			LockTime:      nil,
		}
		playerStates[playerID] = playerState
	}
	
	// Create match state
	matchState := &InMemoryMatchState{
		MatchID:       matchID,
		League:        league,
		Status:        MatchStatusForming,
		CurrentHeat:   0, // No heat started yet
		HeatStatus:    HeatStatusWaiting,
		HeatStartTime: nil,
		HeatEndTime:   nil,
		Players:       playerStates,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	m.states[matchID] = matchState
	
	m.logger.WithFields(logrus.Fields{
		"match_id":     matchID,
		"league":       league,
		"player_count": len(players),
	}).Info("Match state created")
	
	return nil
}

// GetMatchState retrieves a match state
func (m *matchStateManager) GetMatchState(ctx context.Context, matchID uuid.UUID) (*InMemoryMatchState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	state, exists := m.states[matchID]
	if !exists {
		return nil, fmt.Errorf("match state not found for match %s", matchID)
	}
	
	// Return a copy to prevent external modifications
	stateCopy := InMemoryMatchState{
		MatchID:       state.MatchID,
		League:        state.League,
		Status:        state.Status,
		CurrentHeat:   state.CurrentHeat,
		HeatStatus:    state.HeatStatus,
		HeatStartTime: state.HeatStartTime,
		HeatEndTime:   state.HeatEndTime,
		CreatedAt:     state.CreatedAt,
		UpdatedAt:     state.UpdatedAt,
		Players:       make(map[uuid.UUID]*InMemoryPlayer),
	}
	for id, player := range state.Players {
		playerCopy := *player
		stateCopy.Players[id] = &playerCopy
	}
	
	return &stateCopy, nil
}

// UpdateMatchStatus updates the match status
func (m *matchStateManager) UpdateMatchStatus(ctx context.Context, matchID uuid.UUID, status MatchStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	state, exists := m.states[matchID]
	if !exists {
		return fmt.Errorf("match state not found for match %s", matchID)
	}
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	oldStatus := state.Status
	state.Status = status
	state.UpdatedAt = time.Now()
	
	m.logger.WithFields(logrus.Fields{
		"match_id":   matchID,
		"old_status": oldStatus,
		"new_status": status,
	}).Info("Match status updated")
	
	return nil
}

// StartHeat starts a specific heat
func (m *matchStateManager) StartHeat(ctx context.Context, matchID uuid.UUID, heat int) error {
	if heat < 1 || heat > 3 {
		return fmt.Errorf("invalid heat number: %d", heat)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	state, exists := m.states[matchID]
	if !exists {
		return fmt.Errorf("match state not found for match %s", matchID)
	}
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	// Update heat state
	state.CurrentHeat = heat
	state.HeatStatus = HeatStatusCountdown
	now := time.Now()
	state.HeatStartTime = &now
	state.HeatEndTime = nil
	state.UpdatedAt = now
	
	// Reset player states for new heat
	for _, player := range state.Players {
		player.IsAlive = true
		player.HasLocked = false
		player.LockTime = nil
	}
	
	m.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"heat":     heat,
	}).Info("Heat started")
	
	return nil
}

// EndHeat ends the current heat
func (m *matchStateManager) EndHeat(ctx context.Context, matchID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	state, exists := m.states[matchID]
	if !exists {
		return fmt.Errorf("match state not found for match %s", matchID)
	}
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	// Update heat state
	state.HeatStatus = HeatStatusCompleted
	now := time.Now()
	state.HeatEndTime = &now
	state.UpdatedAt = now
	
	// Calculate positions for this heat
	m.calculateHeatPositions(state)
	
	m.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"heat":     state.CurrentHeat,
	}).Info("Heat ended")
	
	// If this was Heat 3, the match is complete
	if state.CurrentHeat == 3 {
		state.Status = MatchStatusCompleted
		m.calculateFinalPositions(state)
	} else {
		// Transition to intermission
		state.HeatStatus = HeatStatusIntermission
	}
	
	return nil
}

// LockPlayerScore locks a player's score for the current heat
func (m *matchStateManager) LockPlayerScore(ctx context.Context, matchID, userID uuid.UUID, score decimal.Decimal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	state, exists := m.states[matchID]
	if !exists {
		return fmt.Errorf("match state not found for match %s", matchID)
	}
	
	state.mu.Lock()
	defer state.mu.Unlock()
	
	// Find player
	var player *InMemoryPlayer
	for id, p := range state.Players {
		if p.UserID != nil && *p.UserID == userID {
			player = p
			break
		} else if id == userID { // For ghost players
			player = p
			break
		}
	}
	
	if player == nil {
		return fmt.Errorf("player not found in match: %s", userID)
	}
	
	if player.HasLocked {
		return fmt.Errorf("player has already locked score for this heat")
	}
	
	if state.HeatStatus != HeatStatusActive {
		return fmt.Errorf("heat is not active")
	}
	
	// Lock the score
	now := time.Now()
	player.HasLocked = true
	player.LockTime = &now
	
	// Set score for current heat
	switch state.CurrentHeat {
	case 1:
		player.Heat1Score = &score
	case 2:
		player.Heat2Score = &score
	case 3:
		player.Heat3Score = &score
	}
	
	// Update total score
	m.calculatePlayerTotalScore(player)
	
	state.UpdatedAt = now
	
	m.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"user_id":  userID,
		"heat":     state.CurrentHeat,
		"score":    score,
	}).Info("Player score locked")
	
	return nil
}

// GetActiveMatches returns all active match IDs
func (m *matchStateManager) GetActiveMatches(ctx context.Context) []uuid.UUID {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var activeMatches []uuid.UUID
	for matchID, state := range m.states {
		if state.Status == MatchStatusInProgress {
			activeMatches = append(activeMatches, matchID)
		}
	}
	
	return activeMatches
}

// RemoveMatchState removes a match state from memory
func (m *matchStateManager) RemoveMatchState(ctx context.Context, matchID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.states, matchID)
	
	m.logger.WithFields(logrus.Fields{
		"match_id": matchID,
	}).Info("Match state removed from memory")
	
	return nil
}

// calculatePlayerTotalScore calculates a player's total score across all heats
func (m *matchStateManager) calculatePlayerTotalScore(player *InMemoryPlayer) {
	total := decimal.Zero
	
	if player.Heat1Score != nil {
		total = total.Add(*player.Heat1Score)
	}
	if player.Heat2Score != nil {
		total = total.Add(*player.Heat2Score)
	}
	if player.Heat3Score != nil {
		total = total.Add(*player.Heat3Score)
	}
	
	player.TotalScore = total
}

// calculateHeatPositions calculates positions for the current heat
func (m *matchStateManager) calculateHeatPositions(state *InMemoryMatchState) {
	// Create slice of players for sorting
	type playerScore struct {
		player *InMemoryPlayer
		score  decimal.Decimal
	}
	
	var players []playerScore
	for _, player := range state.Players {
		var score decimal.Decimal
		switch state.CurrentHeat {
		case 1:
			if player.Heat1Score != nil {
				score = *player.Heat1Score
			}
		case 2:
			if player.Heat2Score != nil {
				score = *player.Heat2Score
			}
		case 3:
			if player.Heat3Score != nil {
				score = *player.Heat3Score
			}
		}
		
		players = append(players, playerScore{
			player: player,
			score:  score,
		})
	}
	
	// Sort by score (descending)
	for i := 0; i < len(players)-1; i++ {
		for j := i + 1; j < len(players); j++ {
			if players[i].score.LessThan(players[j].score) {
				players[i], players[j] = players[j], players[i]
			}
		}
	}
	
	// Assign positions
	for i, ps := range players {
		ps.player.Position = i + 1
	}
}

// calculateFinalPositions calculates final positions based on total scores
func (m *matchStateManager) calculateFinalPositions(state *InMemoryMatchState) {
	// Create slice of players for sorting
	type playerTotal struct {
		player *InMemoryPlayer
		total  decimal.Decimal
	}
	
	var players []playerTotal
	for _, player := range state.Players {
		m.calculatePlayerTotalScore(player)
		players = append(players, playerTotal{
			player: player,
			total:  player.TotalScore,
		})
	}
	
	// Sort by total score (descending), with tiebreaker logic
	for i := 0; i < len(players)-1; i++ {
		for j := i + 1; j < len(players); j++ {
			if m.shouldSwapPlayers(players[i].player, players[j].player) {
				players[i], players[j] = players[j], players[i]
			}
		}
	}
	
	// Assign final positions
	for i, ps := range players {
		ps.player.Position = i + 1
	}
}

// shouldSwapPlayers determines if two players should be swapped in sorting
// Implements tiebreaker logic: Heat 3 → Heat 2 → Heat 1
func (m *matchStateManager) shouldSwapPlayers(p1, p2 *InMemoryPlayer) bool {
	// First, compare total scores
	if p1.TotalScore.GreaterThan(p2.TotalScore) {
		return false // p1 is better
	}
	if p1.TotalScore.LessThan(p2.TotalScore) {
		return true // p2 is better
	}
	
	// Total scores are equal, use tiebreaker
	// Heat 3 tiebreaker
	h1_h3 := decimal.Zero
	h2_h3 := decimal.Zero
	if p1.Heat3Score != nil {
		h1_h3 = *p1.Heat3Score
	}
	if p2.Heat3Score != nil {
		h2_h3 = *p2.Heat3Score
	}
	
	if h1_h3.GreaterThan(h2_h3) {
		return false // p1 is better
	}
	if h1_h3.LessThan(h2_h3) {
		return true // p2 is better
	}
	
	// Heat 2 tiebreaker
	h1_h2 := decimal.Zero
	h2_h2 := decimal.Zero
	if p1.Heat2Score != nil {
		h1_h2 = *p1.Heat2Score
	}
	if p2.Heat2Score != nil {
		h2_h2 = *p2.Heat2Score
	}
	
	if h1_h2.GreaterThan(h2_h2) {
		return false // p1 is better
	}
	if h1_h2.LessThan(h2_h2) {
		return true // p2 is better
	}
	
	// Heat 1 tiebreaker
	h1_h1 := decimal.Zero
	h2_h1 := decimal.Zero
	if p1.Heat1Score != nil {
		h1_h1 = *p1.Heat1Score
	}
	if p2.Heat1Score != nil {
		h2_h1 = *p2.Heat1Score
	}
	
	return h1_h1.LessThan(h2_h1) // p2 is better if their Heat 1 score is higher
}