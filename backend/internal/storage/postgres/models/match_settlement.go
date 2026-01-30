package models

import (
	"time"

	"github.com/google/uuid"
)

// MatchSettlement ensures idempotent settlement (applied exactly once per match)
type MatchSettlement struct {
	MatchID   uuid.UUID `db:"match_id" json:"match_id"`
	SettledAt time.Time `db:"settled_at" json:"settled_at"`
}