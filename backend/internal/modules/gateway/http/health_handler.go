package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/services"
)

// HealthHandler handles health check HTTP endpoints
type HealthHandler struct {
	container *services.Container
	logger    *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(container *services.Container, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		container: container,
		logger:    logger,
	}
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.HealthCheck)
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Perform health checks on all services
	if err := h.container.HealthCheck(ctx); err != nil {
		h.logger.WithError(err).Error("Health check failed")

		response := NewHealthResponse("unhealthy", "nitro-drag-royale", err)

		render.Status(r, http.StatusServiceUnavailable)
		render.Render(w, r, response)
		return
	}

	response := NewHealthResponse("healthy", "nitro-drag-royale", nil)

	render.Status(r, http.StatusOK)
	render.Render(w, r, response)
}
