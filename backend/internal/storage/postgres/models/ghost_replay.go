package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GhostReplay represents historical race performance data used to fill Ghost slots
type GhostReplay struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	SourceMatchID  uuid.UUID       `db:"source_match_id" json:"source_match_id"`
	SourceUserID   uuid.UUID       `db:"source_user_id" json:"source_user_id"`
	League         League          `db:"league" json:"league"`
	DisplayName    string          `db:"display_name" json:"display_name"`
	Heat1Score     decimal.Decimal `db:"heat1_score" json:"heat1_score"`
	Heat2Score     decimal.Decimal `db:"heat2_score" json:"heat2_score"`
	Heat3Score     decimal.Decimal `db:"heat3_score" json:"heat3_score"`
	TotalScore     decimal.Decimal `db:"total_score" json:"total_score"`
	BehavioralData json.RawMessage `db:"behavioral_data" json:"behavioral_data"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

// BehavioralData represents the timing data for Ghost behavior
type BehavioralData struct {
	Heat1LockTime float64 `json:"heat1_lock_time"`
	Heat2LockTime float64 `json:"heat2_lock_time"`
	Heat3LockTime float64 `json:"heat3_lock_time"`
}

// GetBehavioralData parses the behavioral_data JSONB field
func (gr *GhostReplay) GetBehavioralData() (*BehavioralData, error) {
	var data BehavioralData
	err := json.Unmarshal(gr.BehavioralData, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// SetBehavioralData sets the behavioral_data JSONB field
func (gr *GhostReplay) SetBehavioralData(data *BehavioralData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	gr.BehavioralData = jsonData
	return nil
}
