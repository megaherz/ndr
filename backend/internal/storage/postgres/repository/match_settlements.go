package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/megaherz/ndr/internal/storage/postgres/models"
)

// MatchSettlementRepository defines the interface for match settlement data access
type MatchSettlementRepository interface {
	// Create creates a new match settlement record
	Create(ctx context.Context, settlement *models.MatchSettlement) error
	
	// GetByMatchID retrieves a settlement by match ID
	GetByMatchID(ctx context.Context, matchID uuid.UUID) (*models.MatchSettlement, error)
	
	// IsSettled checks if a match has been settled
	IsSettled(ctx context.Context, matchID uuid.UUID) (bool, error)
}

// matchSettlementRepository implements MatchSettlementRepository
type matchSettlementRepository struct {
	db *sqlx.DB
}

// NewMatchSettlementRepository creates a new match settlement repository
func NewMatchSettlementRepository(db *sqlx.DB) MatchSettlementRepository {
	return &matchSettlementRepository{db: db}
}

// Create creates a new match settlement record
func (r *matchSettlementRepository) Create(ctx context.Context, settlement *models.MatchSettlement) error {
	query := `
		INSERT INTO match_settlements (match_id, settled_at)
		VALUES (:match_id, :settled_at)`
	
	_, err := r.db.NamedExecContext(ctx, query, settlement)
	return err
}

// GetByMatchID retrieves a settlement by match ID
func (r *matchSettlementRepository) GetByMatchID(ctx context.Context, matchID uuid.UUID) (*models.MatchSettlement, error) {
	settlement := &models.MatchSettlement{}
	query := `
		SELECT match_id, settled_at
		FROM match_settlements 
		WHERE match_id = $1`
	
	err := r.db.GetContext(ctx, settlement, query, matchID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return settlement, nil
}

// IsSettled checks if a match has been settled
func (r *matchSettlementRepository) IsSettled(ctx context.Context, matchID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM match_settlements WHERE match_id = $1`
	
	err := r.db.GetContext(ctx, &count, query, matchID)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}