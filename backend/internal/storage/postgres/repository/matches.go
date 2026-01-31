package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"github.com/megaherz/ndr/internal/storage/postgres/models"
)

// MatchRepository defines the interface for match data access
type MatchRepository interface {
	// Create creates a new match
	Create(ctx context.Context, match *models.Match) error
	
	// GetByID retrieves a match by ID
	GetByID(ctx context.Context, matchID uuid.UUID) (*models.Match, error)
	
	// UpdateStatus updates the match status
	UpdateStatus(ctx context.Context, matchID uuid.UUID, status string) error
	
	// SetStartTime sets the match start timestamp
	SetStartTime(ctx context.Context, matchID uuid.UUID) error
	
	// SetCompletionTime sets the match completion timestamp
	SetCompletionTime(ctx context.Context, matchID uuid.UUID) error
	
	// GetActiveMatches retrieves all matches that are currently in progress
	GetActiveMatches(ctx context.Context) ([]*models.Match, error)
	
	// GetMatchHistory retrieves match history for a user with pagination
	GetMatchHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Match, error)
	
	// GetLeagueStats retrieves statistics for a league (total matches, avg prize pool, etc.)
	GetLeagueStats(ctx context.Context, league string) (*LeagueStats, error)
}

// LeagueStats represents statistics for a league
type LeagueStats struct {
	League           string          `json:"league"`
	TotalMatches     int64           `json:"total_matches"`
	ActiveMatches    int64           `json:"active_matches"`
	AvgPrizePool     decimal.Decimal `json:"avg_prize_pool"`
	TotalPrizePool   decimal.Decimal `json:"total_prize_pool"`
	TotalRakeAmount  decimal.Decimal `json:"total_rake_amount"`
}

// matchRepository implements MatchRepository
type matchRepository struct {
	db *sqlx.DB
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(db *sqlx.DB) MatchRepository {
	return &matchRepository{db: db}
}

// Create creates a new match
func (r *matchRepository) Create(ctx context.Context, match *models.Match) error {
	query := `
		INSERT INTO matches (id, league, status, live_player_count, ghost_player_count,
		                    prize_pool, rake_amount, crash_seed, crash_seed_hash,
		                    started_at, completed_at, created_at)
		VALUES (:id, :league, :status, :live_player_count, :ghost_player_count,
		        :prize_pool, :rake_amount, :crash_seed, :crash_seed_hash,
		        :started_at, :completed_at, :created_at)`
	
	_, err := r.db.NamedExecContext(ctx, query, match)
	return err
}

// GetByID retrieves a match by ID
func (r *matchRepository) GetByID(ctx context.Context, matchID uuid.UUID) (*models.Match, error) {
	match := &models.Match{}
	query := `
		SELECT id, league, status, live_player_count, ghost_player_count,
		       prize_pool, rake_amount, crash_seed, crash_seed_hash,
		       started_at, completed_at, created_at
		FROM matches 
		WHERE id = $1`
	
	err := r.db.GetContext(ctx, match, query, matchID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return match, nil
}

// UpdateStatus updates the match status
func (r *matchRepository) UpdateStatus(ctx context.Context, matchID uuid.UUID, status string) error {
	query := `UPDATE matches SET status = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, matchID, status)
	return err
}

// SetStartTime sets the match start timestamp
func (r *matchRepository) SetStartTime(ctx context.Context, matchID uuid.UUID) error {
	query := `UPDATE matches SET started_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, matchID)
	return err
}

// SetCompletionTime sets the match completion timestamp
func (r *matchRepository) SetCompletionTime(ctx context.Context, matchID uuid.UUID) error {
	query := `UPDATE matches SET completed_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, matchID)
	return err
}

// GetActiveMatches retrieves all matches that are currently in progress
func (r *matchRepository) GetActiveMatches(ctx context.Context) ([]*models.Match, error) {
	matches := []*models.Match{}
	query := `
		SELECT id, league, status, live_player_count, ghost_player_count,
		       prize_pool, rake_amount, crash_seed, crash_seed_hash,
		       started_at, completed_at, created_at
		FROM matches 
		WHERE status IN ('FORMING', 'IN_PROGRESS')
		ORDER BY created_at ASC`
	
	err := r.db.SelectContext(ctx, &matches, query)
	return matches, err
}

// GetMatchHistory retrieves match history for a user with pagination
func (r *matchRepository) GetMatchHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Match, error) {
	matches := []*models.Match{}
	query := `
		SELECT m.id, m.league, m.status, m.live_player_count, m.ghost_player_count,
		       m.prize_pool, m.rake_amount, m.crash_seed, m.crash_seed_hash,
		       m.started_at, m.completed_at, m.created_at
		FROM matches m
		INNER JOIN match_participants mp ON m.id = mp.match_id
		WHERE mp.user_id = $1 AND mp.is_ghost = FALSE
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3`
	
	err := r.db.SelectContext(ctx, &matches, query, userID, limit, offset)
	return matches, err
}

// GetLeagueStats retrieves statistics for a league
func (r *matchRepository) GetLeagueStats(ctx context.Context, league string) (*LeagueStats, error) {
	stats := &LeagueStats{League: league}
	
	// Get basic match counts and prize pool stats
	query := `
		SELECT 
			COUNT(*) as total_matches,
			COUNT(CASE WHEN status IN ('FORMING', 'IN_PROGRESS') THEN 1 END) as active_matches,
			COALESCE(AVG(prize_pool), 0) as avg_prize_pool,
			COALESCE(SUM(prize_pool), 0) as total_prize_pool,
			COALESCE(SUM(rake_amount), 0) as total_rake_amount
		FROM matches 
		WHERE league = $1`
	
	row := r.db.QueryRowContext(ctx, query, league)
	err := row.Scan(
		&stats.TotalMatches,
		&stats.ActiveMatches,
		&stats.AvgPrizePool,
		&stats.TotalPrizePool,
		&stats.TotalRakeAmount,
	)
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}