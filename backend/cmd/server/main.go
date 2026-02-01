package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/config"
	"github.com/megaherz/ndr/internal/metrics"
	"github.com/megaherz/ndr/internal/modules/gateway/routes"
	"github.com/megaherz/ndr/internal/services"
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

	// Initialize service container with all dependencies
	container, err := services.NewContainer(cfg, logrus.StandardLogger())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize service container")
	}

	// Ensure graceful shutdown of services
	defer func() {
		if err := container.Close(); err != nil {
			logrus.WithError(err).Error("Error closing service container")
		}
	}()

	// Setup HTTP router with all routes and middleware
	r := routes.SetupRoutes(container, logrus.StandardLogger())

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
