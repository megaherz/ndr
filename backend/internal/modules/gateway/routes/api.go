package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"

	httpHandlers "github.com/megaherz/ndr/internal/modules/gateway/http"
	gatewayMiddleware "github.com/megaherz/ndr/internal/modules/gateway/middleware"
	"github.com/megaherz/ndr/internal/services"
)

// SetupRoutes configures and returns the main HTTP router
func SetupRoutes(container *services.Container, logger *logrus.Logger) chi.Router {
	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(gatewayMiddleware.LogrusMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware for Telegram Mini App
	r.Use(gatewayMiddleware.CORS())

	// Initialize handlers
	authHandler := httpHandlers.NewAuthHandler(container.AuthService, logger)
	healthHandler := httpHandlers.NewHealthHandler(container, logger)
	walletHandler := httpHandlers.NewWalletHandler(container.AccountService, logger)
	garageHandler := httpHandlers.NewGarageHandler(container.AccountService, logger)

	// Health check endpoint (outside of API versioning)
	healthHandler.RegisterRoutes(r)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// API root endpoint
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			response := httpHandlers.NewAPIInfoResponse("Nitro Drag Royale API v1", "ready")

			render.Status(r, http.StatusOK)
			render.Render(w, r, response)
		})

		// Authentication routes (no auth required)
		authHandler.RegisterRoutes(r)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			// JWT authentication middleware
			r.Use(gatewayMiddleware.JWTAuth(container.JWTManager, logger))

			// Wallet routes
			walletHandler.RegisterRoutes(r)

			// Garage routes
			garageHandler.RegisterRoutes(r)
		})
	})

	return r
}
