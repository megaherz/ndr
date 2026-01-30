package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"ndr/internal/config"
	"ndr/internal/metrics"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Setup logging
	setupLogging(cfg)

	logrus.WithFields(logrus.Fields{
		"version":     "0.1.0",
		"environment": cfg.Environment,
		"port":        cfg.Port,
	}).Info("Starting Nitro Drag Royale server")

	// Initialize metrics
	metricsInstance := metrics.New()

	// Setup HTTP router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"nitro-drag-royale"}`))
	})

	// API routes (placeholder)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Nitro Drag Royale API v1","status":"ready"}`))
		})
	})

	// Start metrics server
	go func() {
		metricsServer := &http.Server{
			Addr:    cfg.MetricsAddr,
			Handler: metricsInstance.Handler(),
		}

		logrus.WithField("addr", cfg.MetricsAddr).Info("Starting metrics server")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Error("Metrics server failed")
		}
	}()

	// Start main HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logrus.WithField("port", cfg.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("HTTP server failed")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}

// setupLogging configures the logging system
func setupLogging(cfg *config.Config) {
	// Set log level
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Set log format
	if cfg.IsProduction() {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}

	logrus.WithFields(logrus.Fields{
		"level":       level.String(),
		"environment": cfg.Environment,
	}).Info("Logging configured")
}
