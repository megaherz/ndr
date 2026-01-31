package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/modules/auth"
)

// AuthHandler handles authentication HTTP endpoints
type AuthHandler struct {
	authService auth.AuthService
	logger      *logrus.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService auth.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/telegram", h.AuthenticateTelegram)
		r.Post("/refresh", h.RefreshToken)
	})
}

// TelegramAuthRequest represents the request body for Telegram authentication
type TelegramAuthRequest struct {
	InitData string `json:"init_data" validate:"required"`
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthenticateTelegram handles POST /api/v1/auth/telegram
func (h *AuthHandler) AuthenticateTelegram(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req TelegramAuthRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to decode authentication request")

		render.Status(r, http.StatusBadRequest)
		render.Render(w, r, NewErrorResponse("Invalid request body"))
		return
	}

	// Validate required fields
	if req.InitData == "" {
		render.Status(r, http.StatusBadRequest)
		render.Render(w, r, NewErrorResponse("init_data is required"))
		return
	}

	// Authenticate user
	result, err := h.authService.Authenticate(ctx, req.InitData)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Authentication failed")

		render.Status(r, http.StatusUnauthorized)
		render.Render(w, r, NewErrorResponse("Authentication failed"))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":     result.User.ID,
		"telegram_id": result.User.TelegramID,
	}).Info("User authenticated successfully via HTTP")

	// Return success response
	render.Status(r, http.StatusOK)
	render.Render(w, r, NewSuccessResponse(result))
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req RefreshTokenRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to decode refresh token request")

		render.Status(r, http.StatusBadRequest)
		render.Render(w, r, NewErrorResponse("Invalid request body"))
		return
	}

	// Validate required fields
	if req.RefreshToken == "" {
		render.Status(r, http.StatusBadRequest)
		render.Render(w, r, NewErrorResponse("refresh_token is required"))
		return
	}

	// Refresh token
	result, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Token refresh failed")

		render.Status(r, http.StatusUnauthorized)
		render.Render(w, r, NewErrorResponse("Token refresh failed"))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": result.User.ID,
	}).Info("Token refreshed successfully")

	// Return success response
	render.Status(r, http.StatusOK)
	render.Render(w, r, NewSuccessResponse(result))
}
