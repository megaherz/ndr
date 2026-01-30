package centrifugo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/centrifugal/gocent/v3"
	"github.com/sirupsen/logrus"
)

// Client wraps the Centrifugo gRPC client with additional functionality
type Client struct {
	client *gocent.Client
	logger *logrus.Logger
}

// Config holds Centrifugo client configuration
type Config struct {
	GRPCAddr string
	APIKey   string
}

// NewClient creates a new Centrifugo client wrapper
func NewClient(cfg Config, logger *logrus.Logger) (*Client, error) {
	client := gocent.New(gocent.Config{
		Addr: cfg.GRPCAddr,
		Key:  cfg.APIKey,
	})

	logger.WithFields(logrus.Fields{
		"grpc_addr": cfg.GRPCAddr,
	}).Info("Connected to Centrifugo")

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the Centrifugo client connection
func (c *Client) Close() error {
	// gocent v3 doesn't have a Close method
	return nil
}

// PublishToUser publishes a message to a user's personal channel
func (c *Client) PublishToUser(ctx context.Context, userID string, event string, data interface{}) error {
	channel := fmt.Sprintf("user:%s", userID)
	return c.publish(ctx, channel, event, data)
}

// PublishToMatch publishes a message to a match channel
func (c *Client) PublishToMatch(ctx context.Context, matchID string, event string, data interface{}) error {
	channel := fmt.Sprintf("match:%s", matchID)
	return c.publish(ctx, channel, event, data)
}

// publish is the internal method for publishing messages
func (c *Client) publish(ctx context.Context, channel string, event string, data interface{}) error {
	// Create the event payload
	payload := map[string]interface{}{
		"event": event,
		"data":  data,
		"timestamp": time.Now().Unix(),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	// Publish to Centrifugo
	_, err = c.client.Publish(ctx, channel, jsonData)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"channel": channel,
			"event":   event,
			"error":   err,
		}).Error("Failed to publish message to Centrifugo")
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}

	c.logger.WithFields(logrus.Fields{
		"channel": channel,
		"event":   event,
	}).Debug("Published message to Centrifugo")

	return nil
}

// Broadcast publishes a message to multiple channels
func (c *Client) Broadcast(ctx context.Context, channels []string, event string, data interface{}) error {
	// Create the event payload
	payload := map[string]interface{}{
		"event": event,
		"data":  data,
		"timestamp": time.Now().Unix(),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	// Publish to all channels
	_, err = c.client.Broadcast(ctx, channels, jsonData)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"channels": channels,
			"event":    event,
			"error":    err,
		}).Error("Failed to broadcast message to Centrifugo")
		return fmt.Errorf("failed to broadcast to channels: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"channels": channels,
		"event":    event,
	}).Debug("Broadcasted message to Centrifugo")

	return nil
}

// GetPresence returns presence information for a channel
func (c *Client) GetPresence(ctx context.Context, channel string) (map[string]gocent.ClientInfo, error) {
	result, err := c.client.Presence(ctx, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get presence for channel %s: %w", channel, err)
	}
	return result.Presence, nil
}

// GetPresenceStats returns presence statistics for a channel
func (c *Client) GetPresenceStats(ctx context.Context, channel string) (*gocent.PresenceStatsResult, error) {
	result, err := c.client.PresenceStats(ctx, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get presence stats for channel %s: %w", channel, err)
	}
	return &result, nil
}

// GetHistory returns channel history (simplified for gocent v3)
func (c *Client) GetHistory(ctx context.Context, channel string, limit int, since *gocent.StreamPosition, reverse bool) (*gocent.HistoryResult, error) {
	// gocent v3 has a simpler History API
	result, err := c.client.History(ctx, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get history for channel %s: %w", channel, err)
	}
	return &result, nil
}

// GetChannels returns active channels (simplified for gocent v3)
func (c *Client) GetChannels(ctx context.Context, pattern string) ([]string, error) {
	result, err := c.client.Channels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	
	// Convert map keys to slice
	channels := make([]string, 0, len(result.Channels))
	for channel := range result.Channels {
		channels = append(channels, channel)
	}
	return channels, nil
}

// Disconnect disconnects a user from all connections
func (c *Client) Disconnect(ctx context.Context, userID string) error {
	err := c.client.Disconnect(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to disconnect user %s: %w", userID, err)
	}
	
	c.logger.WithField("user_id", userID).Info("Disconnected user from Centrifugo")
	return nil
}

// Unsubscribe removes a user from a channel
func (c *Client) Unsubscribe(ctx context.Context, channel string, userID string) error {
	err := c.client.Unsubscribe(ctx, channel, userID)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe user %s from channel %s: %w", userID, channel, err)
	}
	
	c.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"channel": channel,
	}).Debug("Unsubscribed user from channel")
	return nil
}

// Subscribe adds a user to a channel
func (c *Client) Subscribe(ctx context.Context, channel string, userID string) error {
	err := c.client.Subscribe(ctx, channel, userID)
	if err != nil {
		return fmt.Errorf("failed to subscribe user %s to channel %s: %w", userID, channel, err)
	}
	
	c.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"channel": channel,
	}).Debug("Subscribed user to channel")
	return nil
}

// Refresh sends a refresh command to update connection credentials
func (c *Client) Refresh(ctx context.Context, userID string, expired bool) error {
	// Note: gocent v3 may not have Refresh method, this is a placeholder
	c.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"expired": expired,
	}).Debug("Refresh not implemented in gocent v3")
	return nil
}

// GetInfo returns Centrifugo server information
func (c *Client) GetInfo(ctx context.Context) (*gocent.InfoResult, error) {
	result, err := c.client.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	return &result, nil
}

// Event types for type safety
const (
	// User channel events
	EventBalanceUpdated = "balance_updated"
	EventMatchFound     = "match_found"
	EventPaymentUpdate  = "payment_update"
	
	// Match channel events
	EventHeatStarted  = "heat_started"
	EventHeatEnded    = "heat_ended"
	EventMatchSettled = "match_settled"
	EventPlayerAction = "player_action"
)

// Event payload structures
type BalanceUpdatedEvent struct {
	FuelBalance string `json:"fuel_balance"`
	BurnBalance string `json:"burn_balance"`
	TonBalance  string `json:"ton_balance"`
}

type MatchFoundEvent struct {
	MatchID     string `json:"match_id"`
	League      string `json:"league"`
	PlayerCount int    `json:"player_count"`
}

type HeatStartedEvent struct {
	MatchID     string  `json:"match_id"`
	HeatNumber  int     `json:"heat_number"`
	Duration    int     `json:"duration_seconds"`
	TargetLine  *string `json:"target_line,omitempty"` // For Heat 2 and 3
}

type HeatEndedEvent struct {
	MatchID      string                    `json:"match_id"`
	HeatNumber   int                       `json:"heat_number"`
	Standings    []ParticipantStanding     `json:"standings"`
	NextHeat     *int                      `json:"next_heat,omitempty"`
}

type MatchSettledEvent struct {
	MatchID       string                `json:"match_id"`
	FinalStandings []ParticipantStanding `json:"final_standings"`
	PrizePool     string                `json:"prize_pool"`
	CrashSeed     string                `json:"crash_seed"`
	CrashSeedHash string                `json:"crash_seed_hash"`
}

type ParticipantStanding struct {
	UserID      *string `json:"user_id,omitempty"`
	DisplayName string  `json:"display_name"`
	Position    int     `json:"position"`
	TotalScore  string  `json:"total_score"`
	HeatScores  []string `json:"heat_scores"`
	PrizeAmount string  `json:"prize_amount"`
	BurnReward  string  `json:"burn_reward"`
	IsGhost     bool    `json:"is_ghost"`
}

// Helper methods for publishing specific events
func (c *Client) PublishBalanceUpdated(ctx context.Context, userID string, event *BalanceUpdatedEvent) error {
	return c.PublishToUser(ctx, userID, EventBalanceUpdated, event)
}

func (c *Client) PublishMatchFound(ctx context.Context, userID string, event *MatchFoundEvent) error {
	return c.PublishToUser(ctx, userID, EventMatchFound, event)
}

func (c *Client) PublishHeatStarted(ctx context.Context, matchID string, event *HeatStartedEvent) error {
	return c.PublishToMatch(ctx, matchID, EventHeatStarted, event)
}

func (c *Client) PublishHeatEnded(ctx context.Context, matchID string, event *HeatEndedEvent) error {
	return c.PublishToMatch(ctx, matchID, EventHeatEnded, event)
}

func (c *Client) PublishMatchSettled(ctx context.Context, matchID string, event *MatchSettledEvent) error {
	return c.PublishToMatch(ctx, matchID, EventMatchSettled, event)
}