package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to decode authentication request")
		
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate required fields
	if req.InitData == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "init_data is required")
		return
	}
	
	// Authenticate user
	result, err := h.authService.Authenticate(ctx, req.InitData)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Authentication failed")
		
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication failed")
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"user_id":     result.User.ID,
		"telegram_id": result.User.TelegramID,
	}).Info("User authenticated successfully via HTTP")
	
	// Return success response
	h.writeSuccessResponse(w, result)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse request body
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to decode refresh token request")
		
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate required fields
	if req.RefreshToken == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "refresh_token is required")
		return
	}
	
	// Refresh token
	result, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Token refresh failed")
		
		h.writeErrorResponse(w, http.StatusUnauthorized, "Token refresh failed")
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"user_id": result.User.ID,
	}).Info("Token refreshed successfully")
	
	// Return success response
	h.writeSuccessResponse(w, result)
}

// writeSuccessResponse writes a successful JSON response
func (h *AuthHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := AuthResponse{
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
func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := AuthResponse{
		Success: false,
		Error:   message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to encode error response")
	}
}