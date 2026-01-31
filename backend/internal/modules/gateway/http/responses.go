package http

import (
	"net/http"
	"time"
)

// APIResponse represents a standardized API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// Render implements chi/render.Renderer interface
func (ar *APIResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Set timestamp if not already set
	if ar.Timestamp == "" {
		ar.Timestamp = time.Now().Format(time.RFC3339)
	}
	return nil
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(message string) *APIResponse {
	return &APIResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

// Render implements chi/render.Renderer interface
func (hr *HealthResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if hr.Timestamp == "" {
		hr.Timestamp = time.Now().Format(time.RFC3339)
	}
	return nil
}

// NewHealthResponse creates a health check response
func NewHealthResponse(status, service string, err error) *HealthResponse {
	response := &HealthResponse{
		Status:    status,
		Service:   service,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}

// APIInfoResponse represents the API info response
type APIInfoResponse struct {
	Message   string `json:"message"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Render implements chi/render.Renderer interface
func (air *APIInfoResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if air.Timestamp == "" {
		air.Timestamp = time.Now().Format(time.RFC3339)
	}
	return nil
}

// NewAPIInfoResponse creates an API info response
func NewAPIInfoResponse(message, status string) *APIInfoResponse {
	return &APIInfoResponse{
		Message:   message,
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
