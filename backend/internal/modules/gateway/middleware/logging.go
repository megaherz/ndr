package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

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
