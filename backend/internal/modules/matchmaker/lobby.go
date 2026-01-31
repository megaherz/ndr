package matchmaker

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"backend/internal/modules/gameengine"
	"backend/internal/modules/gateway"
	"backend/internal/modules/gateway/events"
)

// LobbyManager handles lobby formation and management
type LobbyManager interface {
	// FormLobby attempts to form a lobby from the queue
	FormLobby(ctx context.Context, league string) (*Lobby, error)
	
	// CheckTimeout checks if any lobbies have timed out
	CheckTimeout(ctx context.Context) error
	
	// GetActiveLobby returns an active lobby for a user
	GetActiveLobby(ctx context.Context, userID uuid.UUID) (*Lobby, error)
}

// Lobby represents a formed lobby waiting to start a match
type Lobby struct {
	ID          uuid.UUID        `json:"id"`
	League      string           `json:"league"`
	Players     []*LobbyPlayer   `json:"players"`
	Status      LobbyStatus      `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	StartTime   *time.Time       `json:"start_time,omitempty"`
	TimeoutAt   time.Time        `json:"timeout_at"`
}

// LobbyPlayer represents a player in a lobby
type LobbyPlayer struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	IsReady     bool      `json:"is_ready"`
	JoinedAt    time.Time `json:"joined_at"`
}

// LobbyStatus represents the status of a lobby
type LobbyStatus string

const (
	LobbyStatusForming   LobbyStatus = "FORMING"   // Waiting for players to ready up
	LobbyStatusCountdown LobbyStatus = "COUNTDOWN" // Countdown to match start
	LobbyStatusStarted   LobbyStatus = "STARTED"   // Match has started
	LobbyStatusAborted   LobbyStatus = "ABORTED"   // Lobby was cancelled
)

// lobbyManager implements LobbyManager
type lobbyManager struct {
	queueOps      QueueOperations
	gameEngine    gameengine.GameEngineService
	publisher     gateway.CentrifugoPublisher
	activeLobies  map[uuid.UUID]*Lobby // In-memory lobby storage
	userToLobby   map[uuid.UUID]uuid.UUID // User to lobby mapping
	logger        *logrus.Logger
}

// NewLobbyManager creates a new lobby manager
func NewLobbyManager(
	queueOps QueueOperations,
	gameEngine gameengine.GameEngineService,
	publisher gateway.CentrifugoPublisher,
	logger *logrus.Logger,
) LobbyManager {
	return &lobbyManager{
		queueOps:     queueOps,
		gameEngine:   gameEngine,
		publisher:    publisher,
		activeLobies: make(map[uuid.UUID]*Lobby),
		userToLobby:  make(map[uuid.UUID]uuid.UUID),
		logger:       logger,
	}
}

// FormLobby attempts to form a lobby from the queue
func (lm *lobbyManager) FormLobby(ctx context.Context, league string) (*Lobby, error) {
	// Check if we have enough players in the queue
	queueSize, err := lm.queueOps.GetQueueSize(ctx, league)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue size: %w", err)
	}
	
	if queueSize < 10 {
		return nil, fmt.Errorf("not enough players in queue: %d/10", queueSize)
	}
	
	// Pop 10 players from the queue
	queueEntries, err := lm.queueOps.PopPlayersFromQueue(ctx, league, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to pop players from queue: %w", err)
	}
	
	if len(queueEntries) < 10 {
		// Put players back in queue if we didn't get enough
		for _, entry := range queueEntries {
			if addErr := lm.queueOps.AddToQueue(ctx, league, entry); addErr != nil {
				lm.logger.WithFields(logrus.Fields{
					"user_id": entry.UserID,
					"league":  league,
					"error":   addErr,
				}).Error("Failed to re-add player to queue")
			}
		}
		return nil, fmt.Errorf("insufficient players popped from queue: %d/10", len(queueEntries))
	}
	
	// Create lobby
	lobby := &Lobby{
		ID:        uuid.New(),
		League:    league,
		Status:    LobbyStatusForming,
		CreatedAt: time.Now(),
		TimeoutAt: time.Now().Add(getMatchmakingTimeout()),
		Players:   make([]*LobbyPlayer, 0, 10),
	}
	
	// Add players to lobby
	for _, entry := range queueEntries {
		player := &LobbyPlayer{
			UserID:      entry.UserID,
			DisplayName: entry.DisplayName,
			IsReady:     false, // Players need to ready up
			JoinedAt:    time.Now(),
		}
		lobby.Players = append(lobby.Players, player)
		
		// Map user to lobby
		lm.userToLobby[entry.UserID] = lobby.ID
	}
	
	// Store lobby
	lm.activeLobies[lobby.ID] = lobby
	
	lm.logger.WithFields(logrus.Fields{
		"lobby_id":     lobby.ID,
		"league":       league,
		"player_count": len(lobby.Players),
	}).Info("Lobby formed successfully")
	
	// Notify players via Centrifugo that match was found (T059)
	err = lm.publishMatchFoundEvents(ctx, lobby)
	if err != nil {
		lm.logger.WithFields(logrus.Fields{
			"lobby_id": lobby.ID,
			"error":    err,
		}).Error("Failed to publish match found events")
		// Continue anyway - lobby is formed
	}
	
	return lobby, nil
}

// CheckTimeout checks if any lobbies have timed out
func (lm *lobbyManager) CheckTimeout(ctx context.Context) error {
	now := time.Now()
	
	for lobbyID, lobby := range lm.activeLobies {
		if now.After(lobby.TimeoutAt) && lobby.Status == LobbyStatusForming {
			lm.logger.WithFields(logrus.Fields{
				"lobby_id": lobbyID,
				"league":   lobby.League,
			}).Info("Lobby timed out, aborting")
			
			// Abort the lobby
			err := lm.abortLobby(ctx, lobby)
			if err != nil {
				lm.logger.WithFields(logrus.Fields{
					"lobby_id": lobbyID,
					"error":    err,
				}).Error("Failed to abort timed out lobby")
			}
		}
	}
	
	return nil
}

// GetActiveLobby returns an active lobby for a user
func (lm *lobbyManager) GetActiveLobby(ctx context.Context, userID uuid.UUID) (*Lobby, error) {
	lobbyID, exists := lm.userToLobby[userID]
	if !exists {
		return nil, nil // User not in any lobby
	}
	
	lobby, exists := lm.activeLobies[lobbyID]
	if !exists {
		// Clean up stale mapping
		delete(lm.userToLobby, userID)
		return nil, nil
	}
	
	return lobby, nil
}

// abortLobby aborts a lobby and returns players to queue
func (lm *lobbyManager) abortLobby(ctx context.Context, lobby *Lobby) error {
	// Change status
	lobby.Status = LobbyStatusAborted
	
	// Return players to queue
	for _, player := range lobby.Players {
		// Create queue entry
		queueEntry := &QueueEntry{
			UserID:      player.UserID,
			DisplayName: player.DisplayName,
			League:      lobby.League,
			BuyinAmount: LeagueBuyins[lobby.League],
			JoinedAt:    time.Now(),
		}
		
		// Add back to queue
		err := lm.queueOps.AddToQueue(ctx, lobby.League, queueEntry)
		if err != nil {
			lm.logger.WithFields(logrus.Fields{
				"user_id": player.UserID,
				"league":  lobby.League,
				"error":   err,
			}).Error("Failed to return player to queue after lobby abort")
		}
		
		// Clean up user mapping
		delete(lm.userToLobby, player.UserID)
	}
	
	// Remove lobby
	delete(lm.activeLobies, lobby.ID)
	
	// TODO: Notify players via Centrifugo that lobby was aborted
	
	return nil
}

// startMatch starts a match from a ready lobby
func (lm *lobbyManager) startMatch(ctx context.Context, lobby *Lobby) error {
	// Validate all players are ready
	for _, player := range lobby.Players {
		if !player.IsReady {
			return fmt.Errorf("not all players are ready")
		}
	}
	
	// Change lobby status
	lobby.Status = LobbyStatusCountdown
	now := time.Now()
	lobby.StartTime = &now
	
	// TODO: Create match via GameEngineService
	// matchID, err := lm.gameEngine.CreateMatch(ctx, lobby)
	// if err != nil {
	//     return fmt.Errorf("failed to create match: %w", err)
	// }
	
	// Change status to started
	lobby.Status = LobbyStatusStarted
	
	// Clean up lobby (players are now in match)
	for _, player := range lobby.Players {
		delete(lm.userToLobby, player.UserID)
	}
	delete(lm.activeLobies, lobby.ID)
	
	lm.logger.WithFields(logrus.Fields{
		"lobby_id": lobby.ID,
		"league":   lobby.League,
	}).Info("Match started from lobby")
	
	return nil
}

// publishMatchFoundEvents publishes match_found events to all players in the lobby
func (lm *lobbyManager) publishMatchFoundEvents(ctx context.Context, lobby *Lobby) error {
	// Calculate total buyin amount for prize pool
	totalBuyin := decimal.Zero
	for _, player := range lobby.Players {
		totalBuyin = totalBuyin.Add(LeagueBuyins[lobby.League])
	}
	
	// Calculate prize pool (after 8% rake)
	rakeAmount := totalBuyin.Mul(decimal.NewFromFloat(0.08)).Truncate(2)
	prizePool := totalBuyin.Sub(rakeAmount)
	
	// Create match found event
	matchFoundEvent := &events.MatchFoundEvent{
		MatchID:         lobby.ID, // Using lobby ID as match ID for now
		League:          lobby.League,
		PlayerCount:     len(lobby.Players),
		BuyinAmount:     LeagueBuyins[lobby.League],
		PrizePool:       prizePool,
		CountdownStart:  time.Now().Add(5 * time.Second), // 5 seconds from now
	}
	
	// Publish to each player's personal channel
	for _, player := range lobby.Players {
		err := lm.publisher.PublishToUser(ctx, player.UserID, events.EventMatchFound, matchFoundEvent)
		if err != nil {
			lm.logger.WithFields(logrus.Fields{
				"lobby_id": lobby.ID,
				"user_id":  player.UserID,
				"error":    err,
			}).Error("Failed to publish match found event to player")
			// Continue with other players
		}
	}
	
	lm.logger.WithFields(logrus.Fields{
		"lobby_id":     lobby.ID,
		"league":       lobby.League,
		"player_count": len(lobby.Players),
		"prize_pool":   prizePool,
	}).Info("Published match found events to all players")
	
	return nil
}

// getMatchmakingTimeout returns the matchmaking timeout from environment variable
func getMatchmakingTimeout() time.Duration {
	timeoutStr := os.Getenv("MATCHMAKING_TIMEOUT_SECONDS")
	if timeoutStr == "" {
		return 60 * time.Second // Default 60 seconds
	}
	
	timeoutSeconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return 60 * time.Second // Default on error
	}
	
	return time.Duration(timeoutSeconds) * time.Second
}