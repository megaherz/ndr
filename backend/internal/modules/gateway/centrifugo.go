package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"ndr/internal/centrifugo"
)

// CentrifugoPublisher handles publishing events to Centrifugo channels
type CentrifugoPublisher interface {
	// PublishToUser publishes an event to a user's personal channel
	PublishToUser(ctx context.Context, userID uuid.UUID, eventType string, data interface{}) error
	
	// PublishToMatch publishes an event to a match channel
	PublishToMatch(ctx context.Context, matchID uuid.UUID, eventType string, data interface{}) error
	
	// PublishToUsers publishes an event to multiple user channels
	PublishToUsers(ctx context.Context, userIDs []uuid.UUID, eventType string, data interface{}) error
	
	// BroadcastToChannel publishes an event to a specific channel
	BroadcastToChannel(ctx context.Context, channel string, eventType string, data interface{}) error
}

// centrifugoPublisher implements CentrifugoPublisher
type centrifugoPublisher struct {
	client *centrifugo.Client
	logger *logrus.Logger
}

// NewCentrifugoPublisher creates a new Centrifugo publisher
func NewCentrifugoPublisher(client *centrifugo.Client, logger *logrus.Logger) CentrifugoPublisher {
	return &centrifugoPublisher{
		client: client,
		logger: logger,
	}
}

// EventMessage represents a standardized event message structure
type EventMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// PublishToUser publishes an event to a user's personal channel
func (p *centrifugoPublisher) PublishToUser(ctx context.Context, userID uuid.UUID, eventType string, data interface{}) error {
	channel := fmt.Sprintf("user:%s", userID.String())
	return p.BroadcastToChannel(ctx, channel, eventType, data)
}

// PublishToMatch publishes an event to a match channel
func (p *centrifugoPublisher) PublishToMatch(ctx context.Context, matchID uuid.UUID, eventType string, data interface{}) error {
	channel := fmt.Sprintf("match:%s", matchID.String())
	return p.BroadcastToChannel(ctx, channel, eventType, data)
}

// PublishToUsers publishes an event to multiple user channels
func (p *centrifugoPublisher) PublishToUsers(ctx context.Context, userIDs []uuid.UUID, eventType string, data interface{}) error {
	// Prepare the event message once
	message, err := p.prepareEventMessage(eventType, data)
	if err != nil {
		return fmt.Errorf("failed to prepare event message: %w", err)
	}
	
	// Publish to each user channel
	for _, userID := range userIDs {
		channel := fmt.Sprintf("user:%s", userID.String())
		if err := p.publishMessage(ctx, channel, message); err != nil {
			// Log error but continue with other users
			p.logger.WithFields(logrus.Fields{
				"user_id":    userID,
				"channel":    channel,
				"event_type": eventType,
				"error":      err,
			}).Error("Failed to publish event to user")
		}
	}
	
	return nil
}

// BroadcastToChannel publishes an event to a specific channel
func (p *centrifugoPublisher) BroadcastToChannel(ctx context.Context, channel string, eventType string, data interface{}) error {
	message, err := p.prepareEventMessage(eventType, data)
	if err != nil {
		return fmt.Errorf("failed to prepare event message: %w", err)
	}
	
	return p.publishMessage(ctx, channel, message)
}

// prepareEventMessage creates a standardized event message
func (p *centrifugoPublisher) prepareEventMessage(eventType string, data interface{}) (*EventMessage, error) {
	message := &EventMessage{
		Type:      eventType,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
	}
	
	return message, nil
}

// publishMessage publishes a message to a specific channel
func (p *centrifugoPublisher) publishMessage(ctx context.Context, channel string, message *EventMessage) error {
	// Serialize the message
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal event message: %w", err)
	}
	
	// Publish to Centrifugo
	err = p.client.Publish(ctx, channel, messageData)
	if err != nil {
		p.logger.WithFields(logrus.Fields{
			"channel":    channel,
			"event_type": message.Type,
			"error":      err,
		}).Error("Failed to publish event to Centrifugo")
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}
	
	p.logger.WithFields(logrus.Fields{
		"channel":    channel,
		"event_type": message.Type,
	}).Debug("Successfully published event to Centrifugo")
	
	return nil
}

// getCurrentTimestamp returns the current Unix timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}