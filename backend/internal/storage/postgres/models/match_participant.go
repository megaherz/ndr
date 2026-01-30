package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MatchParticipant represents a player's participation in a match
type MatchParticipant struct {
	ID                int64            `db:"id" json:"id"`
	MatchID           uuid.UUID        `db:"match_id" json:"match_id"`
	UserID            *uuid.UUID       `db:"user_id" json:"user_id,omitempty"`
	IsGhost           bool             `db:"is_ghost" json:"is_ghost"`
	GhostReplayID     *uuid.UUID       `db:"ghost_replay_id" json:"ghost_replay_id,omitempty"`
	PlayerDisplayName string           `db:"player_display_name" json:"player_display_name"`
	BuyinAmount       decimal.Decimal  `db:"buyin_amount" json:"buyin_amount"`
	Heat1Score        *decimal.Decimal `db:"heat1_score" json:"heat1_score,omitempty"`
	Heat2Score        *decimal.Decimal `db:"heat2_score" json:"heat2_score,omitempty"`
	Heat3Score        *decimal.Decimal `db:"heat3_score" json:"heat3_score,omitempty"`
	TotalScore        *decimal.Decimal `db:"total_score" json:"total_score,omitempty"`
	FinalPosition     *int             `db:"final_position" json:"final_position,omitempty"`
	PrizeAmount       decimal.Decimal  `db:"prize_amount" json:"prize_amount"`
	BurnReward        decimal.Decimal  `db:"burn_reward" json:"burn_reward"`
	CreatedAt         time.Time        `db:"created_at" json:"created_at"`
}

// CalculateTotalScore calculates the total score from individual heat scores
func (mp *MatchParticipant) CalculateTotalScore() decimal.Decimal {
	total := decimal.Zero
	if mp.Heat1Score != nil {
		total = total.Add(*mp.Heat1Score)
	}
	if mp.Heat2Score != nil {
		total = total.Add(*mp.Heat2Score)
	}
	if mp.Heat3Score != nil {
		total = total.Add(*mp.Heat3Score)
	}
	return total
}

// IsLivePlayer returns true if this is a live player (not a Ghost)
func (mp *MatchParticipant) IsLivePlayer() bool {
	return !mp.IsGhost && mp.UserID != nil
}
