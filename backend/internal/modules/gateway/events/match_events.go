package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Event types for match-related events
const (
	EventMatchFound    = "match_found"
	EventHeatStarted   = "heat_started"
	EventHeatEnded     = "heat_ended"
	EventMatchSettled  = "match_settled"
	EventBalanceUpdated = "balance_updated"
)

// MatchFoundEvent is published to user:{user_id} when a match is found
type MatchFoundEvent struct {
	MatchID         uuid.UUID `json:"match_id"`
	League          string    `json:"league"`
	PlayerCount     int       `json:"player_count"`
	BuyinAmount     decimal.Decimal `json:"buyin_amount"`
	PrizePool       decimal.Decimal `json:"prize_pool"`
	CountdownStart  time.Time `json:"countdown_start"`
}

// HeatStartedEvent is published to match:{match_id} when a heat begins
type HeatStartedEvent struct {
	MatchID     uuid.UUID `json:"match_id"`
	Heat        int       `json:"heat"`        // 1, 2, or 3
	Duration    int       `json:"duration"`    // Heat duration in seconds (25)
	StartTime   time.Time `json:"start_time"`  // When the heat started
	TargetLine  *decimal.Decimal `json:"target_line,omitempty"` // Speed to beat (Heat 2 & 3 only)
	Participants []ParticipantInfo `json:"participants"`
}

// HeatEndedEvent is published to match:{match_id} when a heat ends
type HeatEndedEvent struct {
	MatchID      uuid.UUID `json:"match_id"`
	Heat         int       `json:"heat"`         // 1, 2, or 3
	EndTime      time.Time `json:"end_time"`     // When the heat ended
	ActualDuration float64 `json:"actual_duration"` // Actual duration in seconds
	FinalSpeed   decimal.Decimal `json:"final_speed"`  // Speed at heat end
	Results      []HeatResult `json:"results"`      // All participant results
	Standings    []StandingEntry `json:"standings"`   // Current standings after this heat
}

// MatchSettledEvent is published to match:{match_id} when the match is complete
type MatchSettledEvent struct {
	MatchID         uuid.UUID `json:"match_id"`
	League          string    `json:"league"`
	CompletedAt     time.Time `json:"completed_at"`
	FinalStandings  []FinalStanding `json:"final_standings"`
	PrizeDistribution []PrizeEntry `json:"prize_distribution"`
	CrashSeed       string    `json:"crash_seed"`       // Revealed crash seed data
	CrashSeedHash   string    `json:"crash_seed_hash"`  // Original hash for verification
}

// BalanceUpdatedEvent is published to user:{user_id} when balance changes
type BalanceUpdatedEvent struct {
	UserID       uuid.UUID `json:"user_id"`
	TONBalance   decimal.Decimal `json:"ton_balance"`
	FuelBalance  decimal.Decimal `json:"fuel_balance"`
	BurnBalance  decimal.Decimal `json:"burn_balance"`
	Changes      BalanceChanges `json:"changes"`
	Reason       string    `json:"reason"`       // "match_settlement", "deposit", etc.
	ReferenceID  *uuid.UUID `json:"reference_id,omitempty"` // Match ID, payment ID, etc.
}

// Supporting data structures

// ParticipantInfo represents basic participant information
type ParticipantInfo struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"` // Null for ghosts
	DisplayName string     `json:"display_name"`
	IsGhost     bool       `json:"is_ghost"`
	Position    int        `json:"position"`          // Starting position (1-10)
}

// HeatResult represents a participant's result in a heat
type HeatResult struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"` // Null for ghosts
	DisplayName string     `json:"display_name"`
	IsGhost     bool       `json:"is_ghost"`
	Score       *decimal.Decimal `json:"score"`       // Null if crashed (score = 0)
	LockTime    *float64   `json:"lock_time,omitempty"` // When they locked (seconds into heat)
	Crashed     bool       `json:"crashed"`         // True if score = 0
}

// StandingEntry represents a participant's standing after a heat
type StandingEntry struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"` // Null for ghosts
	DisplayName string     `json:"display_name"`
	IsGhost     bool       `json:"is_ghost"`
	Position    int        `json:"position"`          // Current position (1-10)
	TotalScore  decimal.Decimal `json:"total_score"`
	Heat1Score  *decimal.Decimal `json:"heat1_score,omitempty"`
	Heat2Score  *decimal.Decimal `json:"heat2_score,omitempty"`
	Heat3Score  *decimal.Decimal `json:"heat3_score,omitempty"`
	PositionDelta int      `json:"position_delta"`   // Change from previous heat (↑ +1, ↓ -1, = 0)
}

// FinalStanding represents final match results
type FinalStanding struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"` // Null for ghosts
	DisplayName string     `json:"display_name"`
	IsGhost     bool       `json:"is_ghost"`
	FinalPosition int      `json:"final_position"`    // 1-10
	TotalScore  decimal.Decimal `json:"total_score"`
	Heat1Score  decimal.Decimal `json:"heat1_score"`
	Heat2Score  decimal.Decimal `json:"heat2_score"`
	Heat3Score  decimal.Decimal `json:"heat3_score"`
	PrizeAmount decimal.Decimal `json:"prize_amount"`  // FUEL won
	BurnReward  decimal.Decimal `json:"burn_reward"`   // BURN earned
}

// PrizeEntry represents prize distribution
type PrizeEntry struct {
	Position    int             `json:"position"`
	UserID      *uuid.UUID      `json:"user_id,omitempty"` // Null for ghosts
	DisplayName string          `json:"display_name"`
	IsGhost     bool            `json:"is_ghost"`
	PrizeAmount decimal.Decimal `json:"prize_amount"`
	BurnReward  decimal.Decimal `json:"burn_reward"`
}

// BalanceChanges represents changes to user balances
type BalanceChanges struct {
	TONDelta  decimal.Decimal `json:"ton_delta"`
	FuelDelta decimal.Decimal `json:"fuel_delta"`
	BurnDelta decimal.Decimal `json:"burn_delta"`
}