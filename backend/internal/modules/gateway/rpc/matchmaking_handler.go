package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"backend/internal/modules/matchmaker"
)

// MatchmakingHandler handles matchmaking RPC requests
type MatchmakingHandler struct {
	matchmakerService matchmaker.MatchmakerService
	logger            *logrus.Logger
}

// NewMatchmakingHandler creates a new matchmaking RPC handler
func NewMatchmakingHandler(matchmakerService matchmaker.MatchmakerService, logger *logrus.Logger) *MatchmakingHandler {
	return &MatchmakingHandler{
		matchmakerService: matchmakerService,
		logger:            logger,
	}
}

// JoinMatchmakingRequest represents the request to join matchmaking
type JoinMatchmakingRequest struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	League      string `json:"league"`
}

// JoinMatchmakingResponse represents the response from joining matchmaking
type JoinMatchmakingResponse struct {
	Success     bool                        `json:"success"`
	QueueStatus *matchmaker.QueueStatus     `json:"queue_status,omitempty"`
	Error       string                      `json:"error,omitempty"`
}

// CancelMatchmakingRequest represents the request to cancel matchmaking
type CancelMatchmakingRequest struct {
	UserID string `json:"user_id"`
}

// CancelMatchmakingResponse represents the response from cancelling matchmaking
type CancelMatchmakingResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// HandleJoinMatchmaking handles the matchmaking.join RPC call
func (h *MatchmakingHandler) HandleJoinMatchmaking(ctx context.Context, data []byte) ([]byte, error) {
	var req JoinMatchmakingRequest
	if err := json.Unmarshal(data, &req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to unmarshal join matchmaking request")
		
		return h.errorResponse("Invalid request format")
	}
	
	// Validate request
	if req.UserID == "" {
		return h.errorResponse("user_id is required")
	}
	if req.DisplayName == "" {
		return h.errorResponse("display_name is required")
	}
	if req.League == "" {
		return h.errorResponse("league is required")
	}
	
	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return h.errorResponse("Invalid user_id format")
	}
	
	// Validate league
	validLeagues := map[string]bool{
		"ROOKIE":   true,
		"STREET":   true,
		"PRO":      true,
		"TOP_FUEL": true,
	}
	if !validLeagues[req.League] {
		return h.errorResponse("Invalid league")
	}
	
	// Join matchmaking queue
	queueStatus, err := h.matchmakerService.JoinQueue(ctx, userID, req.DisplayName, req.League)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id":      userID,
			"display_name": req.DisplayName,
			"league":       req.League,
			"error":        err,
		}).Error("Failed to join matchmaking queue")
		
		return h.errorResponse(fmt.Sprintf("Failed to join queue: %s", err.Error()))
	}
	
	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"display_name": req.DisplayName,
		"league":       req.League,
		"position":     queueStatus.Position,
		"queue_size":   queueStatus.QueueSize,
	}).Info("User joined matchmaking queue via RPC")
	
	// Return success response
	response := JoinMatchmakingResponse{
		Success:     true,
		QueueStatus: queueStatus,
	}
	
	return json.Marshal(response)
}

// HandleCancelMatchmaking handles the matchmaking.cancel RPC call
func (h *MatchmakingHandler) HandleCancelMatchmaking(ctx context.Context, data []byte) ([]byte, error) {
	var req CancelMatchmakingRequest
	if err := json.Unmarshal(data, &req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to unmarshal cancel matchmaking request")
		
		return h.errorResponse("Invalid request format")
	}
	
	// Validate request
	if req.UserID == "" {
		return h.errorResponse("user_id is required")
	}
	
	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return h.errorResponse("Invalid user_id format")
	}
	
	// Cancel matchmaking
	err = h.matchmakerService.CancelQueue(ctx, userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to cancel matchmaking queue")
		
		return h.errorResponse(fmt.Sprintf("Failed to cancel queue: %s", err.Error()))
	}
	
	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("User cancelled matchmaking queue via RPC")
	
	// Return success response
	response := CancelMatchmakingResponse{
		Success: true,
	}
	
	return json.Marshal(response)
}

// errorResponse creates an error response
func (h *MatchmakingHandler) errorResponse(message string) ([]byte, error) {
	response := JoinMatchmakingResponse{
		Success: false,
		Error:   message,
	}
	
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal error response: %w", err)
	}
	
	return data, nil
}