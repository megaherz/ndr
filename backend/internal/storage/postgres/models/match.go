package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Match represents a single 10-player race session
type Match struct {
	ID               uuid.UUID       `db:"id" json:"id"`
	League           League          `db:"league" json:"league"`
	Status           MatchStatus     `db:"status" json:"status"`
	LivePlayerCount  int             `db:"live_player_count" json:"live_player_count"`
	GhostPlayerCount int             `db:"ghost_player_count" json:"ghost_player_count"`
	PrizePool        decimal.Decimal `db:"prize_pool" json:"prize_pool"`
	RakeAmount       decimal.Decimal `db:"rake_amount" json:"rake_amount"`
	CrashSeed        string          `db:"crash_seed" json:"crash_seed"`
	CrashSeedHash    string          `db:"crash_seed_hash" json:"crash_seed_hash"`
	StartedAt        *time.Time      `db:"started_at" json:"started_at,omitempty"`
	CompletedAt      *time.Time      `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
}

// League represents the different racing leagues
type League string

const (
	LeagueRookie  League = "ROOKIE"
	LeagueStreet  League = "STREET"
	LeaguePro     League = "PRO"
	LeagueTopFuel League = "TOP_FUEL"
)

// MatchStatus represents the current state of a match
type MatchStatus string

const (
	MatchStatusForming     MatchStatus = "FORMING"
	MatchStatusInProgress  MatchStatus = "IN_PROGRESS"
	MatchStatusCompleted   MatchStatus = "COMPLETED"
	MatchStatusAborted     MatchStatus = "ABORTED"
)