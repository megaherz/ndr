package matchmaker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// QueueEntry represents a player in the matchmaking queue
type QueueEntry struct {
	UserID      uuid.UUID       `json:"user_id"`
	DisplayName string          `json:"display_name"`
	League      string          `json:"league"`
	BuyinAmount decimal.Decimal `json:"buyin_amount"`
	JoinedAt    time.Time       `json:"joined_at"`
}

// QueueOperations handles Redis queue operations for matchmaking
type QueueOperations interface {
	// AddToQueue adds a player to the matchmaking queue for a specific league
	AddToQueue(ctx context.Context, league string, entry *QueueEntry) error
	
	// RemoveFromQueue removes a player from the matchmaking queue
	RemoveFromQueue(ctx context.Context, league string, userID uuid.UUID) error
	
	// GetQueueSize returns the current queue size for a league
	GetQueueSize(ctx context.Context, league string) (int64, error)
	
	// PopPlayersFromQueue removes and returns up to N players from the queue
	PopPlayersFromQueue(ctx context.Context, league string, count int) ([]*QueueEntry, error)
	
	// PeekQueue returns the first N players in the queue without removing them
	PeekQueue(ctx context.Context, league string, count int) ([]*QueueEntry, error)
	
	// IsUserInQueue checks if a user is currently in any queue
	IsUserInQueue(ctx context.Context, userID uuid.UUID) (bool, string, error)
	
	// GetQueuePosition returns the position of a user in the queue (0-based)
	GetQueuePosition(ctx context.Context, league string, userID uuid.UUID) (int64, error)
}

// redisQueueOperations implements QueueOperations using Redis
type redisQueueOperations struct {
	client *redis.Client
}

// NewQueueOperations creates a new Redis-based queue operations handler
func NewQueueOperations(client *redis.Client) QueueOperations {
	return &redisQueueOperations{client: client}
}

// getQueueKey returns the Redis key for a league queue
func (q *redisQueueOperations) getQueueKey(league string) string {
	return fmt.Sprintf("matchmaking:queue:%s", league)
}

// getUserQueueKey returns the Redis key for tracking which queue a user is in
func (q *redisQueueOperations) getUserQueueKey(userID uuid.UUID) string {
	return fmt.Sprintf("matchmaking:user:%s", userID.String())
}

// AddToQueue adds a player to the matchmaking queue for a specific league
func (q *redisQueueOperations) AddToQueue(ctx context.Context, league string, entry *QueueEntry) error {
	// Serialize the queue entry
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal queue entry: %w", err)
	}
	
	// Use a transaction to ensure atomicity
	pipe := q.client.TxPipeline()
	
	// Add to the league queue (FIFO using RPUSH)
	queueKey := q.getQueueKey(league)
	pipe.RPush(ctx, queueKey, data)
	
	// Track which queue the user is in
	userKey := q.getUserQueueKey(entry.UserID)
	pipe.Set(ctx, userKey, league, time.Hour) // Expire after 1 hour as safety
	
	// Execute the transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}
	
	return nil
}

// RemoveFromQueue removes a player from the matchmaking queue
func (q *redisQueueOperations) RemoveFromQueue(ctx context.Context, league string, userID uuid.UUID) error {
	queueKey := q.getQueueKey(league)
	userKey := q.getUserQueueKey(userID)
	
	// Get all entries in the queue
	entries, err := q.client.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get queue entries: %w", err)
	}
	
	// Find and remove the user's entry
	for _, entryData := range entries {
		var entry QueueEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			continue // Skip invalid entries
		}
		
		if entry.UserID == userID {
			// Use transaction to remove from both queue and user tracking
			pipe := q.client.TxPipeline()
			pipe.LRem(ctx, queueKey, 1, entryData)
			pipe.Del(ctx, userKey)
			
			_, err = pipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to remove from queue: %w", err)
			}
			return nil
		}
	}
	
	// User not found in queue, still clean up user tracking
	q.client.Del(ctx, userKey)
	return nil
}

// GetQueueSize returns the current queue size for a league
func (q *redisQueueOperations) GetQueueSize(ctx context.Context, league string) (int64, error) {
	queueKey := q.getQueueKey(league)
	size, err := q.client.LLen(ctx, queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}
	return size, nil
}

// PopPlayersFromQueue removes and returns up to N players from the queue
func (q *redisQueueOperations) PopPlayersFromQueue(ctx context.Context, league string, count int) ([]*QueueEntry, error) {
	queueKey := q.getQueueKey(league)
	
	// Use transaction to pop multiple entries atomically
	pipe := q.client.TxPipeline()
	var cmds []*redis.StringCmd
	
	for i := 0; i < count; i++ {
		cmd := pipe.LPop(ctx, queueKey)
		cmds = append(cmds, cmd)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to pop from queue: %w", err)
	}
	
	var entries []*QueueEntry
	for _, cmd := range cmds {
		entryData, err := cmd.Result()
		if err == redis.Nil {
			break // No more entries
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get popped entry: %w", err)
		}
		
		var entry QueueEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			continue // Skip invalid entries
		}
		
		entries = append(entries, &entry)
		
		// Clean up user tracking
		userKey := q.getUserQueueKey(entry.UserID)
		q.client.Del(ctx, userKey)
	}
	
	return entries, nil
}

// PeekQueue returns the first N players in the queue without removing them
func (q *redisQueueOperations) PeekQueue(ctx context.Context, league string, count int) ([]*QueueEntry, error) {
	queueKey := q.getQueueKey(league)
	
	entryDataList, err := q.client.LRange(ctx, queueKey, 0, int64(count-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to peek queue: %w", err)
	}
	
	var entries []*QueueEntry
	for _, entryData := range entryDataList {
		var entry QueueEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			continue // Skip invalid entries
		}
		entries = append(entries, &entry)
	}
	
	return entries, nil
}

// IsUserInQueue checks if a user is currently in any queue
func (q *redisQueueOperations) IsUserInQueue(ctx context.Context, userID uuid.UUID) (bool, string, error) {
	userKey := q.getUserQueueKey(userID)
	
	league, err := q.client.Get(ctx, userKey).Result()
	if err == redis.Nil {
		return false, "", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to check user queue status: %w", err)
	}
	
	return true, league, nil
}

// GetQueuePosition returns the position of a user in the queue (0-based)
func (q *redisQueueOperations) GetQueuePosition(ctx context.Context, league string, userID uuid.UUID) (int64, error) {
	queueKey := q.getQueueKey(league)
	
	entries, err := q.client.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return -1, fmt.Errorf("failed to get queue entries: %w", err)
	}
	
	for i, entryData := range entries {
		var entry QueueEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			continue // Skip invalid entries
		}
		
		if entry.UserID == userID {
			return int64(i), nil
		}
	}
	
	return -1, nil // User not found in queue
}