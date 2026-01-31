package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"ndr/internal/modules/account"
)

// GarageHandler handles garage-related HTTP endpoints
type GarageHandler struct {
	accountService account.AccountService
	logger         *logrus.Logger
}

// NewGarageHandler creates a new garage handler
func NewGarageHandler(accountService account.AccountService, logger *logrus.Logger) *GarageHandler {
	return &GarageHandler{
		accountService: accountService,
		logger:         logger,
	}
}

// RegisterRoutes registers garage routes
func (h *GarageHandler) RegisterRoutes(r chi.Router) {
	r.Route("/garage", func(r chi.Router) {
		r.Get("/", h.GetGarageInfo)
	})
}

// GarageResponse represents the garage information response
type GarageResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetGarageInfo handles GET /api/v1/garage
func (h *GarageHandler) GetGarageInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get user ID from context (set by authentication middleware)
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to get user ID from context")
		
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	// Get wallet information
	walletInfo, err := h.accountService.GetWallet(ctx, userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get wallet information")
		
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get garage information")
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"fuel_balance": walletInfo.FuelBalance,
		"burn_balance": walletInfo.BurnBalance,
	}).Debug("Retrieved garage information")
	
	// Return garage information
	h.writeSuccessResponse(w, walletInfo)
}

// getUserIDFromContext extracts user ID from the request context
func (h *GarageHandler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	// In a real implementation, this would extract the user ID from JWT claims
	// stored in the context by the authentication middleware
	// For now, we'll return a placeholder
	
	// Check if user_id is set in context (by auth middleware)
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID format in context")
	}
	
	return userID, nil
}

// writeSuccessResponse writes a successful JSON response
func (h *GarageHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := GarageResponse{
		Success: true,
		Data:    data,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to encode success response")
	}
}

// writeErrorResponse writes an error JSON response
func (h *GarageHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := GarageResponse{
		Success: false,
		Error:   message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to encode error response")
	}
}