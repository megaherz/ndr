package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/modules/account"
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

// GetGarageInfo handles GET /api/v1/garage
func (h *GarageHandler) GetGarageInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by authentication middleware)
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to get user ID from context")

		render.Status(r, http.StatusUnauthorized)
		render.Render(w, r, NewErrorResponse("Authentication required"))
		return
	}

	// Get wallet information
	walletInfo, err := h.accountService.GetWallet(ctx, userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get wallet information")

		render.Status(r, http.StatusInternalServerError)
		render.Render(w, r, NewErrorResponse("Failed to get garage information"))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"fuel_balance": walletInfo.FuelBalance,
		"burn_balance": walletInfo.BurnBalance,
	}).Debug("Retrieved garage information")

	// Return garage information
	render.Status(r, http.StatusOK)
	render.Render(w, r, NewSuccessResponse(walletInfo))
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
