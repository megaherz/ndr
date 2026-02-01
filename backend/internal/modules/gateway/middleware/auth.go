package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/auth"
	httpHandlers "github.com/megaherz/ndr/internal/modules/gateway/http"
)

// Context key types to avoid collisions
type contextKey string

const (
	userIDKey     contextKey = "user_id"
	telegramIDKey contextKey = "telegram_id"
	tokenTypeKey  contextKey = "token_type"
)

// JWTAuth creates a JWT authentication middleware
func JWTAuth(jwtManager *auth.JWTManager, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Debug("Missing Authorization header")
				render.Status(r, http.StatusUnauthorized)
				render.Render(w, r, httpHandlers.NewErrorResponse("Authorization header required"))
				return
			}

			// Check for Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				logger.Debug("Invalid Authorization header format")
				render.Status(r, http.StatusUnauthorized)
				render.Render(w, r, httpHandlers.NewErrorResponse("Invalid authorization format"))
				return
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				logger.Debug("Empty token in Authorization header")
				render.Status(r, http.StatusUnauthorized)
				render.Render(w, r, httpHandlers.NewErrorResponse("Token required"))
				return
			}

			// Validate token
			claims, err := jwtManager.ValidateAppToken(tokenString)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err,
				}).Debug("Invalid JWT token")
				render.Status(r, http.StatusUnauthorized)
				render.Render(w, r, httpHandlers.NewErrorResponse("Invalid token"))
				return
			}

			// Add user information to context
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, telegramIDKey, claims.TelegramID)
			ctx = context.WithValue(ctx, tokenTypeKey, claims.TokenType)

			// Log successful authentication
			logger.WithFields(logrus.Fields{
				"user_id":     claims.UserID,
				"telegram_id": claims.TelegramID,
				"path":        r.URL.Path,
			}).Debug("User authenticated via JWT")

			// Continue to next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
