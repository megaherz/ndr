package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"backend/internal/modules/gameengine"
)

// MatchHandler handles match-related RPC requests
type MatchHandler struct {
	earnPointsService gameengine.EarnPointsService
	gameEngineService gameengine.GameEngineService
	logger            *logrus.Logger
}

// NewMatchHandler creates a new match RPC handler
func NewMatchHandler(
	earnPointsService gameengine.EarnPointsService,
	gameEngineService gameengine.GameEngineService,
	logger *logrus.Logger,
) *MatchHandler {
	return &MatchHandler{
		earnPointsService: earnPointsService,
		gameEngineService: gameEngineService,
		logger:            logger,
	}
}

// EarnPointsRequest represents the request to earn points (lock score)
type EarnPointsRequest struct {
	MatchID string `json:"match_id"`
	UserID  string `json:"user_id"`
	Score   string `json:"score"` // Decimal as string for precision
}

// EarnPointsResponse represents the response from earning points
type EarnPointsResponse struct {
	Success bool                           `json:"success"`
	Result  *gameengine.EarnPointsResult   `json:"result,omitempty"`
	Error   string                         `json:"error,omitempty"`
}

// GetMatchInfoRequest represents the request to get match information
type GetMatchInfoRequest struct {
	MatchID string `json:"match_id"`
	UserID  string `json:"user_id"`
}

// GetMatchInfoResponse represents the response with match information
type GetMatchInfoResponse struct {
	Success   bool                        `json:"success"`
	HeatInfo  *gameengine.HeatInfo        `json:"heat_info,omitempty"`
	Error     string                      `json:"error,omitempty"`
}

// HandleEarnPoints handles the match.earn_points RPC call
func (h *MatchHandler) HandleEarnPoints(ctx context.Context, data []byte) ([]byte, error) {
	var req EarnPointsRequest
	if err := json.Unmarshal(data, &req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to unmarshal earn points request")
		
		return h.errorResponse("Invalid request format")
	}
	
	// Validate request
	if req.MatchID == "" {
		return h.errorResponse("match_id is required")
	}
	if req.UserID == "" {
		return h.errorResponse("user_id is required")
	}
	if req.Score == "" {
		return h.errorResponse("score is required")
	}
	
	// Parse match ID
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		return h.errorResponse("Invalid match_id format")
	}
	
	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return h.errorResponse("Invalid user_id format")
	}
	
	// Parse score
	score, err := decimal.NewFromString(req.Score)
	if err != nil {
		return h.errorResponse("Invalid score format")
	}
	
	// Lock the score
	result, err := h.earnPointsService.LockScore(ctx, matchID, userID, score)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"user_id":  userID,
			"score":    score,
			"error":    err,
		}).Error("Failed to lock player score")
		
		return h.errorResponse(fmt.Sprintf("Failed to lock score: %s", err.Error()))
	}
	
	h.logger.WithFields(logrus.Fields{
		"match_id":    matchID,
		"user_id":     userID,
		"score":       score,
		"heat":        result.Heat,
		"total_score": result.TotalScore,
		"position":    result.Position,
	}).Info("Player earned points via RPC")
	
	// Return success response
	response := EarnPointsResponse{
		Success: true,
		Result:  result,
	}
	
	return json.Marshal(response)
}

// HandleGetMatchInfo handles requests for current match information
func (h *MatchHandler) HandleGetMatchInfo(ctx context.Context, data []byte) ([]byte, error) {
	var req GetMatchInfoRequest
	if err := json.Unmarshal(data, &req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to unmarshal get match info request")
		
		return h.errorResponse("Invalid request format")
	}
	
	// Validate request
	if req.MatchID == "" {
		return h.errorResponse("match_id is required")
	}
	
	// Parse match ID
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		return h.errorResponse("Invalid match_id format")
	}
	
	// Get heat information
	heatInfo, err := h.earnPointsService.GetCurrentHeatInfo(ctx, matchID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"match_id": matchID,
			"error":    err,
		}).Error("Failed to get heat info")
		
		return h.errorResponse(fmt.Sprintf("Failed to get match info: %s", err.Error()))
	}
	
	// Return success response
	response := GetMatchInfoResponse{
		Success:  true,
		HeatInfo: heatInfo,
	}
	
	return json.Marshal(response)
}

// errorResponse creates an error response for earn points
func (h *MatchHandler) errorResponse(message string) ([]byte, error) {
	response := EarnPointsResponse{
		Success: false,
		Error:   message,
	}
	
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal error response: %w", err)
	}
	
	return data, nil
}