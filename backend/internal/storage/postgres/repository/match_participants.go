package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"github.com/megaherz/ndr/internal/storage/postgres/models"
)

// MatchParticipantRepository defines the interface for match participant data access
type MatchParticipantRepository interface {
	// Create creates a new match participant
	Create(ctx context.Context, participant *models.MatchParticipant) error
	
	// CreateBatch creates multiple match participants in a transaction
	CreateBatch(ctx context.Context, participants []*models.MatchParticipant) error
	
	// GetByMatchID retrieves all participants for a match
	GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error)
	
	// GetByMatchAndUser retrieves a specific participant in a match
	GetByMatchAndUser(ctx context.Context, matchID, userID uuid.UUID) (*models.MatchParticipant, error)
	
	// UpdateHeatScore updates a participant's score for a specific heat
	UpdateHeatScore(ctx context.Context, matchID, userID uuid.UUID, heat int, score decimal.Decimal) error
	
	// UpdateTotalScore updates a participant's total score
	UpdateTotalScore(ctx context.Context, matchID, userID uuid.UUID, totalScore decimal.Decimal) error
	
	// SetFinalPosition sets the final position for a participant
	SetFinalPosition(ctx context.Context, matchID, userID uuid.UUID, position int) error
	
	// SetPrizeAmount sets the prize amount for a participant
	SetPrizeAmount(ctx context.Context, matchID, userID uuid.UUID, prizeAmount decimal.Decimal) error
	
	// SetBurnReward sets the BURN reward for a participant
	SetBurnReward(ctx context.Context, matchID, userID uuid.UUID, burnReward decimal.Decimal) error
	
	// GetLiveParticipants retrieves only live (non-ghost) participants for a match
	GetLiveParticipants(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error)
	
	// GetGhostParticipants retrieves only ghost participants for a match
	GetGhostParticipants(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error)
	
	// GetStandings retrieves participants ordered by total score (for standings)
	GetStandings(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error)
	
	// GetUserStats retrieves statistics for a user across all matches
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserStats, error)
}

// UserStats represents statistics for a user across all matches
type UserStats struct {
	UserID          uuid.UUID       `json:"user_id"`
	TotalMatches    int64           `json:"total_matches"`
	TotalWins       int64           `json:"total_wins"`      // 1st place finishes
	TotalPodiums    int64           `json:"total_podiums"`   // Top 3 finishes
	TotalEarnings   decimal.Decimal `json:"total_earnings"`  // Prize money won
	TotalBurnEarned decimal.Decimal `json:"total_burn_earned"` // BURN rewards
	AvgPosition     float64         `json:"avg_position"`
	BestPosition    int             `json:"best_position"`
	WorstPosition   int             `json:"worst_position"`
}

// matchParticipantRepository implements MatchParticipantRepository
type matchParticipantRepository struct {
	db *sqlx.DB
}

// NewMatchParticipantRepository creates a new match participant repository
func NewMatchParticipantRepository(db *sqlx.DB) MatchParticipantRepository {
	return &matchParticipantRepository{db: db}
}

// Create creates a new match participant
func (r *matchParticipantRepository) Create(ctx context.Context, participant *models.MatchParticipant) error {
	query := `
		INSERT INTO match_participants (match_id, user_id, is_ghost, ghost_replay_id,
		                               player_display_name, buyin_amount, heat1_score,
		                               heat2_score, heat3_score, total_score,
		                               final_position, prize_amount, burn_reward, created_at)
		VALUES (:match_id, :user_id, :is_ghost, :ghost_replay_id,
		        :player_display_name, :buyin_amount, :heat1_score,
		        :heat2_score, :heat3_score, :total_score,
		        :final_position, :prize_amount, :burn_reward, :created_at)`
	
	_, err := r.db.NamedExecContext(ctx, query, participant)
	return err
}

// CreateBatch creates multiple match participants in a transaction
func (r *matchParticipantRepository) CreateBatch(ctx context.Context, participants []*models.MatchParticipant) error {
	if len(participants) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	query := `
		INSERT INTO match_participants (match_id, user_id, is_ghost, ghost_replay_id,
		                               player_display_name, buyin_amount, heat1_score,
		                               heat2_score, heat3_score, total_score,
		                               final_position, prize_amount, burn_reward, created_at)
		VALUES (:match_id, :user_id, :is_ghost, :ghost_replay_id,
		        :player_display_name, :buyin_amount, :heat1_score,
		        :heat2_score, :heat3_score, :total_score,
		        :final_position, :prize_amount, :burn_reward, :created_at)`
	
	for _, participant := range participants {
		_, err := tx.NamedExecContext(ctx, query, participant)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

// GetByMatchID retrieves all participants for a match
func (r *matchParticipantRepository) GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error) {
	participants := []*models.MatchParticipant{}
	query := `
		SELECT match_id, user_id, is_ghost, ghost_replay_id, player_display_name,
		       buyin_amount, heat1_score, heat2_score, heat3_score, total_score,
		       final_position, prize_amount, burn_reward, created_at
		FROM match_participants 
		WHERE match_id = $1
		ORDER BY created_at ASC`
	
	err := r.db.SelectContext(ctx, &participants, query, matchID)
	return participants, err
}

// GetByMatchAndUser retrieves a specific participant in a match
func (r *matchParticipantRepository) GetByMatchAndUser(ctx context.Context, matchID, userID uuid.UUID) (*models.MatchParticipant, error) {
	participant := &models.MatchParticipant{}
	query := `
		SELECT match_id, user_id, is_ghost, ghost_replay_id, player_display_name,
		       buyin_amount, heat1_score, heat2_score, heat3_score, total_score,
		       final_position, prize_amount, burn_reward, created_at
		FROM match_participants 
		WHERE match_id = $1 AND user_id = $2`
	
	err := r.db.GetContext(ctx, participant, query, matchID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return participant, nil
}

// UpdateHeatScore updates a participant's score for a specific heat
func (r *matchParticipantRepository) UpdateHeatScore(ctx context.Context, matchID, userID uuid.UUID, heat int, score decimal.Decimal) error {
	var query string
	switch heat {
	case 1:
		query = `UPDATE match_participants SET heat1_score = $3 WHERE match_id = $1 AND user_id = $2`
	case 2:
		query = `UPDATE match_participants SET heat2_score = $3 WHERE match_id = $1 AND user_id = $2`
	case 3:
		query = `UPDATE match_participants SET heat3_score = $3 WHERE match_id = $1 AND user_id = $2`
	default:
		return sql.ErrNoRows
	}
	
	_, err := r.db.ExecContext(ctx, query, matchID, userID, score)
	return err
}

// UpdateTotalScore updates a participant's total score
func (r *matchParticipantRepository) UpdateTotalScore(ctx context.Context, matchID, userID uuid.UUID, totalScore decimal.Decimal) error {
	query := `UPDATE match_participants SET total_score = $3 WHERE match_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, matchID, userID, totalScore)
	return err
}

// SetFinalPosition sets the final position for a participant
func (r *matchParticipantRepository) SetFinalPosition(ctx context.Context, matchID, userID uuid.UUID, position int) error {
	query := `UPDATE match_participants SET final_position = $3 WHERE match_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, matchID, userID, position)
	return err
}

// SetPrizeAmount sets the prize amount for a participant
func (r *matchParticipantRepository) SetPrizeAmount(ctx context.Context, matchID, userID uuid.UUID, prizeAmount decimal.Decimal) error {
	query := `UPDATE match_participants SET prize_amount = $3 WHERE match_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, matchID, userID, prizeAmount)
	return err
}

// SetBurnReward sets the BURN reward for a participant
func (r *matchParticipantRepository) SetBurnReward(ctx context.Context, matchID, userID uuid.UUID, burnReward decimal.Decimal) error {
	query := `UPDATE match_participants SET burn_reward = $3 WHERE match_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, matchID, userID, burnReward)
	return err
}

// GetLiveParticipants retrieves only live (non-ghost) participants for a match
func (r *matchParticipantRepository) GetLiveParticipants(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error) {
	participants := []*models.MatchParticipant{}
	query := `
		SELECT match_id, user_id, is_ghost, ghost_replay_id, player_display_name,
		       buyin_amount, heat1_score, heat2_score, heat3_score, total_score,
		       final_position, prize_amount, burn_reward, created_at
		FROM match_participants 
		WHERE match_id = $1 AND is_ghost = FALSE
		ORDER BY created_at ASC`
	
	err := r.db.SelectContext(ctx, &participants, query, matchID)
	return participants, err
}

// GetGhostParticipants retrieves only ghost participants for a match
func (r *matchParticipantRepository) GetGhostParticipants(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error) {
	participants := []*models.MatchParticipant{}
	query := `
		SELECT match_id, user_id, is_ghost, ghost_replay_id, player_display_name,
		       buyin_amount, heat1_score, heat2_score, heat3_score, total_score,
		       final_position, prize_amount, burn_reward, created_at
		FROM match_participants 
		WHERE match_id = $1 AND is_ghost = TRUE
		ORDER BY created_at ASC`
	
	err := r.db.SelectContext(ctx, &participants, query, matchID)
	return participants, err
}

// GetStandings retrieves participants ordered by total score (for standings)
func (r *matchParticipantRepository) GetStandings(ctx context.Context, matchID uuid.UUID) ([]*models.MatchParticipant, error) {
	participants := []*models.MatchParticipant{}
	query := `
		SELECT match_id, user_id, is_ghost, ghost_replay_id, player_display_name,
		       buyin_amount, heat1_score, heat2_score, heat3_score, total_score,
		       final_position, prize_amount, burn_reward, created_at
		FROM match_participants 
		WHERE match_id = $1
		ORDER BY 
			total_score DESC NULLS LAST,
			heat3_score DESC NULLS LAST,
			heat2_score DESC NULLS LAST,
			heat1_score DESC NULLS LAST,
			created_at ASC`
	
	err := r.db.SelectContext(ctx, &participants, query, matchID)
	return participants, err
}

// GetUserStats retrieves statistics for a user across all matches
func (r *matchParticipantRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*UserStats, error) {
	stats := &UserStats{UserID: userID}
	
	query := `
		SELECT 
			COUNT(*) as total_matches,
			COUNT(CASE WHEN final_position = 1 THEN 1 END) as total_wins,
			COUNT(CASE WHEN final_position <= 3 THEN 1 END) as total_podiums,
			COALESCE(SUM(prize_amount), 0) as total_earnings,
			COALESCE(SUM(burn_reward), 0) as total_burn_earned,
			COALESCE(AVG(final_position), 0) as avg_position,
			COALESCE(MIN(final_position), 0) as best_position,
			COALESCE(MAX(final_position), 0) as worst_position
		FROM match_participants 
		WHERE user_id = $1 AND is_ghost = FALSE AND final_position IS NOT NULL`
	
	row := r.db.QueryRowContext(ctx, query, userID)
	err := row.Scan(
		&stats.TotalMatches,
		&stats.TotalWins,
		&stats.TotalPodiums,
		&stats.TotalEarnings,
		&stats.TotalBurnEarned,
		&stats.AvgPosition,
		&stats.BestPosition,
		&stats.WorstPosition,
	)
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}