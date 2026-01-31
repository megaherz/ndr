package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"github.com/megaherz/ndr/internal/config"
	"github.com/megaherz/ndr/internal/metrics"
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

	// TODO: Initialize database connection and proper auth service
	// For now, we'll create a temporary auth endpoint to test connectivity

	// Setup HTTP router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(LogrusMiddleware(logrus.StandardLogger()))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware for Telegram Mini App
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"nitro-drag-royale"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Nitro Drag Royale API v1","status":"ready"}`))
		})
		
		// Temporary auth endpoint for testing
		r.Post("/auth/telegram", func(w http.ResponseWriter, r *http.Request) {
			// Log the request for debugging
			logrus.Info("Received auth request")
			
			// Read and parse request body
			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				logrus.WithError(err).Error("Failed to decode request body")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"Invalid request body"}`))
				return
			}
			
			logrus.WithField("request", requestBody).Info("Auth request body")
			
			// Extract user data from request if available
			userID := "temp-user-id"
			telegramID := int64(123456789)
			username := "testuser"
			firstName := "Test"
			lastName := "User"
			
			if user, ok := requestBody["user"].(map[string]interface{}); ok {
				if id, ok := user["id"].(float64); ok {
					telegramID = int64(id)
				}
				if name, ok := user["username"].(string); ok && name != "" {
					username = name
				}
				if name, ok := user["first_name"].(string); ok && name != "" {
					firstName = name
				}
				if name, ok := user["last_name"].(string); ok && name != "" {
					lastName = name
				}
			}
			
			// Generate response with actual user data
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"user": map[string]interface{}{
						"id":                    userID,
						"telegram_id":           telegramID,
						"telegram_username":     username,
						"telegram_first_name":   firstName,
						"telegram_last_name":    lastName,
						"created_at":           time.Now().Format(time.RFC3339),
						"updated_at":           time.Now().Format(time.RFC3339),
					},
					"tokens": map[string]interface{}{
						"app_token":        "temp.jwt.token.for." + username,
						"centrifugo_token": "temp.centrifugo.token.for." + username,
						"expires_at":       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
					},
				},
				"success":   true,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
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

// LogrusMiddleware creates a logrus-based logging middleware for Chi
func LogrusMiddleware(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start)

				fields := logrus.Fields{
					"method":     r.Method,
					"path":       r.URL.Path,
					"status":     ww.Status(),
					"bytes":      ww.BytesWritten(),
					"duration":   duration,
					"request_id": middleware.GetReqID(r.Context()),
					"remote_ip":  r.RemoteAddr,
					"user_agent": r.UserAgent(),
				}

				// Add query parameters if present
				if r.URL.RawQuery != "" {
					fields["query"] = r.URL.RawQuery
				}

				// Log level based on status code
				status := ww.Status()
				switch {
				case status >= 500:
					logger.WithFields(fields).Error("HTTP request completed with server error")
				case status >= 400:
					logger.WithFields(fields).Warn("HTTP request completed with client error")
				default:
					logger.WithFields(fields).Info("HTTP request completed")
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
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
