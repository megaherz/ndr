package matchmaker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"backend/internal/constants"
	"backend/internal/modules/account"
	"backend/internal/modules/gateway"
	"backend/internal/modules/gateway/events"
)

// MatchmakerService handles matchmaking operations
type MatchmakerService interface {
	// JoinQueue adds a player to the matchmaking queue
	JoinQueue(ctx context.Context, userID uuid.UUID, displayName, league string) (*QueueStatus, error)
	
	// CancelQueue removes a player from the matchmaking queue
	CancelQueue(ctx context.Context, userID uuid.UUID) error
	
	// GetQueueStatus returns the current queue status for a user
	GetQueueStatus(ctx context.Context, userID uuid.UUID) (*QueueStatus, error)
	
	// GetQueueInfo returns information about a league's queue
	GetQueueInfo(ctx context.Context, league string) (*QueueInfo, error)
	
	// StartMatchmakingWorker starts the background worker that forms lobbies
	StartMatchmakingWorker(ctx context.Context) error
}

// QueueStatus represents a player's status in the matchmaking queue
type QueueStatus struct {
	InQueue      bool      `json:"in_queue"`
	League       string    `json:"league,omitempty"`
	Position     int64     `json:"position,omitempty"`     // 0-based position in queue
	QueueSize    int64     `json:"queue_size,omitempty"`   // Total players in queue
	EstimatedWait int      `json:"estimated_wait,omitempty"` // Estimated wait time in seconds
	JoinedAt     time.Time `json:"joined_at,omitempty"`
}

// QueueInfo represents information about a league's queue
type QueueInfo struct {
	League       string `json:"league"`
	QueueSize    int64  `json:"queue_size"`
	PlayersNeeded int   `json:"players_needed"` // Players needed for next match
	AvgWaitTime  int    `json:"avg_wait_time"`  // Average wait time in seconds
}

// Use league constants from constants package
var LeagueBuyins = constants.LeagueBuyins

// matchmakerService implements MatchmakerService
type matchmakerService struct {
	queueOps       QueueOperations
	accountService account.AccountService
	publisher      gateway.CentrifugoPublisher
	logger         *logrus.Logger
}

// NewMatchmakerService creates a new matchmaker service
func NewMatchmakerService(
	queueOps QueueOperations,
	accountService account.AccountService,
	publisher gateway.CentrifugoPublisher,
	logger *logrus.Logger,
) MatchmakerService {
	return &matchmakerService{
		queueOps:       queueOps,
		accountService: accountService,
		publisher:      publisher,
		logger:         logger,
	}
}

// JoinQueue adds a player to the matchmaking queue
func (s *matchmakerService) JoinQueue(ctx context.Context, userID uuid.UUID, displayName, league string) (*QueueStatus, error) {
	// Validate league
	buyinAmount, exists := LeagueBuyins[league]
	if !exists {
		return nil, fmt.Errorf("invalid league: %s", league)
	}
	
	// Check if user is already in a queue
	inQueue, currentLeague, err := s.queueOps.IsUserInQueue(ctx, userID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to check if user is in queue")
		return nil, fmt.Errorf("failed to check queue status: %w", err)
	}
	
	if inQueue {
		return nil, fmt.Errorf("user is already in queue for league %s", currentLeague)
	}
	
	// Check if user has sufficient balance
	hasSufficientBalance, err := s.accountService.HasSufficientBalance(ctx, userID, constants.CurrencyFUEL, buyinAmount)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"league":  league,
			"amount":  buyinAmount,
			"error":   err,
		}).Error("Failed to check user balance")
		return nil, fmt.Errorf("failed to check balance: %w", err)
	}
	
	if !hasSufficientBalance {
		return nil, fmt.Errorf("insufficient FUEL balance for %s league (need %s FUEL)", league, buyinAmount.String())
	}
	
	// Create queue entry
	queueEntry := &QueueEntry{
		UserID:      userID,
		DisplayName: displayName,
		League:      league,
		BuyinAmount: buyinAmount,
		JoinedAt:    time.Now(),
	}
	
	// Add to queue
	err = s.queueOps.AddToQueue(ctx, league, queueEntry)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id":     userID,
			"league":      league,
			"buyin_amount": buyinAmount,
			"error":       err,
		}).Error("Failed to add user to queue")
		return nil, fmt.Errorf("failed to join queue: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"display_name": displayName,
		"league":       league,
		"buyin_amount": buyinAmount,
	}).Info("User joined matchmaking queue")
	
	// Get queue status
	return s.GetQueueStatus(ctx, userID)
}

// CancelQueue removes a player from the matchmaking queue
func (s *matchmakerService) CancelQueue(ctx context.Context, userID uuid.UUID) error {
	// Check if user is in a queue
	inQueue, league, err := s.queueOps.IsUserInQueue(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check queue status: %w", err)
	}
	
	if !inQueue {
		return fmt.Errorf("user is not in any queue")
	}
	
	// Remove from queue
	err = s.queueOps.RemoveFromQueue(ctx, league, userID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"league":  league,
			"error":   err,
		}).Error("Failed to remove user from queue")
		return fmt.Errorf("failed to cancel queue: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"league":  league,
	}).Info("User cancelled matchmaking queue")
	
	return nil
}

// GetQueueStatus returns the current queue status for a user
func (s *matchmakerService) GetQueueStatus(ctx context.Context, userID uuid.UUID) (*QueueStatus, error) {
	// Check if user is in a queue
	inQueue, league, err := s.queueOps.IsUserInQueue(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check queue status: %w", err)
	}
	
	if !inQueue {
		return &QueueStatus{
			InQueue: false,
		}, nil
	}
	
	// Get queue position
	position, err := s.queueOps.GetQueuePosition(ctx, league, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue position: %w", err)
	}
	
	// Get queue size
	queueSize, err := s.queueOps.GetQueueSize(ctx, league)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue size: %w", err)
	}
	
	// Calculate estimated wait time (rough estimate based on position)
	estimatedWait := s.calculateEstimatedWaitTime(position, queueSize)
	
	return &QueueStatus{
		InQueue:       true,
		League:        league,
		Position:      position,
		QueueSize:     queueSize,
		EstimatedWait: estimatedWait,
		JoinedAt:      time.Now(), // TODO: Get actual join time from queue entry
	}, nil
}

// GetQueueInfo returns information about a league's queue
func (s *matchmakerService) GetQueueInfo(ctx context.Context, league string) (*QueueInfo, error) {
	// Validate league
	if _, exists := LeagueBuyins[league]; !exists {
		return nil, fmt.Errorf("invalid league: %s", league)
	}
	
	// Get queue size
	queueSize, err := s.queueOps.GetQueueSize(ctx, league)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue size: %w", err)
	}
	
	// Calculate players needed for next match
	playersNeeded := 10 - (int(queueSize) % 10)
	if playersNeeded == 10 && queueSize > 0 {
		playersNeeded = 0 // Queue has exact multiples of 10
	}
	
	// Calculate average wait time (simplified)
	avgWaitTime := s.calculateAverageWaitTime(queueSize)
	
	return &QueueInfo{
		League:        league,
		QueueSize:     queueSize,
		PlayersNeeded: playersNeeded,
		AvgWaitTime:   avgWaitTime,
	}, nil
}

// StartMatchmakingWorker starts the background worker that forms lobbies
func (s *matchmakerService) StartMatchmakingWorker(ctx context.Context) error {
	s.logger.Info("Starting matchmaking worker")
	
	// TODO: Implement the actual matchmaking worker
	// This would run in a goroutine and periodically check queues
	// to form lobbies when enough players are available
	
	go func() {
		ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Matchmaking worker stopped")
				return
			case <-ticker.C:
				// Check each league for lobby formation
				for league := range LeagueBuyins {
					err := s.checkAndFormLobby(ctx, league)
					if err != nil {
						s.logger.WithFields(logrus.Fields{
							"league": league,
							"error":  err,
						}).Error("Failed to check/form lobby")
					}
				}
			}
		}
	}()
	
	return nil
}

// calculateEstimatedWaitTime calculates estimated wait time based on queue position
func (s *matchmakerService) calculateEstimatedWaitTime(position, queueSize int64) int {
	if position == 0 {
		return 5 // First in queue, should match soon
	}
	
	// Rough estimate: assume matches form every 30 seconds
	matchesAhead := position / 10
	return int(matchesAhead * 30)
}

// calculateAverageWaitTime calculates average wait time for a queue
func (s *matchmakerService) calculateAverageWaitTime(queueSize int64) int {
	if queueSize == 0 {
		return 0
	}
	
	// Simplified calculation: more players = faster matches
	if queueSize >= 10 {
		return 10 // Should match quickly
	}
	
	// Estimate based on how many more players needed
	playersNeeded := 10 - queueSize
	return int(playersNeeded * 3) // 3 seconds per player needed
}

// checkAndFormLobby checks if a lobby can be formed for a league
func (s *matchmakerService) checkAndFormLobby(ctx context.Context, league string) error {
	// Get queue size
	queueSize, err := s.queueOps.GetQueueSize(ctx, league)
	if err != nil {
		return err
	}
	
	// Need at least 10 players to form a lobby
	if queueSize < 10 {
		return nil
	}
	
	// TODO: Implement actual lobby formation
	// This would:
	// 1. Pop 10 players from the queue
	// 2. Create a match
	// 3. Notify players via Centrifugo
	// 4. Start the match countdown
	
	s.logger.WithFields(logrus.Fields{
		"league":     league,
		"queue_size": queueSize,
	}).Debug("Could form lobby (not implemented yet)")
	
	return nil
}