package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/shopspring/decimal"

	"github.com/megaherz/ndr/internal/constants"
	"github.com/megaherz/ndr/internal/modules/account"
	"github.com/megaherz/ndr/internal/storage/postgres/models"
	"github.com/megaherz/ndr/internal/storage/postgres/repository"
)

// GarageResponse represents the garage API response
type GarageResponse struct {
	User    GarageUser     `json:"user"`
	Wallet  GarageWallet   `json:"wallet"`
	Leagues []GarageLeague `json:"leagues"`
}

// GarageUser represents user information in garage response
type GarageUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// GarageWallet represents wallet information in garage response
type GarageWallet struct {
	FuelBalance          string `json:"fuel_balance"`
	BurnBalance          string `json:"burn_balance"`
	RookieRacesCompleted int    `json:"rookie_races_completed"`
}

// GarageLeague represents league information in garage response
type GarageLeague struct {
	Name              string  `json:"name"`
	Buyin             string  `json:"buyin"`
	Available         bool    `json:"available"`
	UnavailableReason *string `json:"unavailable_reason,omitempty"`
}

// GarageHandler handles garage-related HTTP endpoints
type GarageHandler struct {
	accountService account.AccountService
	userRepo       repository.UserRepository
	logger         *logrus.Logger
}

// NewGarageHandler creates a new garage handler
func NewGarageHandler(accountService account.AccountService, userRepo repository.UserRepository, logger *logrus.Logger) *GarageHandler {
	return &GarageHandler{
		accountService: accountService,
		userRepo:       userRepo,
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

	// Get user information
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Error("Failed to get user information")

		render.Status(r, http.StatusInternalServerError)
		render.Render(w, r, NewErrorResponse("Failed to get user information"))
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

	// Build garage response
	garageResponse := &GarageResponse{
		User: GarageUser{
			ID:          user.ID.String(),
			DisplayName: getDisplayName(user),
		},
		Wallet: GarageWallet{
			FuelBalance:          walletInfo.FuelBalance.String(),
			BurnBalance:          walletInfo.BurnBalance.String(),
			RookieRacesCompleted: walletInfo.RookieRacesCompleted,
		},
		Leagues: buildLeaguesList(walletInfo),
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"fuel_balance": walletInfo.FuelBalance,
		"burn_balance": walletInfo.BurnBalance,
	}).Debug("Retrieved garage information")

	// Return garage information
	render.Status(r, http.StatusOK)
	render.Render(w, r, NewSuccessResponse(garageResponse))
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

// getDisplayName creates a display name from user information
func getDisplayName(user *models.User) string {
	if user.TelegramUsername != nil && *user.TelegramUsername != "" {
		return *user.TelegramUsername
	}
	
	displayName := user.TelegramFirstName
	if user.TelegramLastName != nil && *user.TelegramLastName != "" {
		displayName += " " + *user.TelegramLastName
	}
	
	return displayName
}

// buildLeaguesList creates the leagues array with availability status
func buildLeaguesList(walletInfo *account.WalletInfo) []GarageLeague {
	leagues := []GarageLeague{
		{
			Name:      "ROOKIE",
			Buyin:     constants.LeagueBuyins["ROOKIE"].String(),
			Available: true,
		},
		{
			Name:      "STREET", 
			Buyin:     constants.LeagueBuyins["STREET"].String(),
			Available: true,
		},
		{
			Name:      "PRO",
			Buyin:     constants.LeagueBuyins["PRO"].String(),
			Available: true,
		},
		{
			Name:      "TOP_FUEL",
			Buyin:     constants.LeagueBuyins["TOP_FUEL"].String(),
			Available: true,
		},
	}

	// Check availability based on wallet state
	for i := range leagues {
		league := &leagues[i]
		buyin, _ := decimal.NewFromString(league.Buyin)

		// Check if user has sufficient balance
		if walletInfo.FuelBalance.LessThan(buyin) {
			league.Available = false
			reason := "INSUFFICIENT_BALANCE"
			league.UnavailableReason = &reason
			continue
		}

		// Check rookie race limit
		if league.Name == "ROOKIE" && walletInfo.RookieRacesCompleted >= 3 {
			league.Available = false
			reason := "ROOKIE_LIMIT_REACHED"
			league.UnavailableReason = &reason
		}
	}

	return leagues
}
