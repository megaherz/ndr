// Code generated manually from matchmaking.proto. DO NOT EDIT.
// Source: matchmaking.proto

package matchmaking

import (
	"encoding/json"
	"fmt"
)

// JoinMatchmakingRequest represents a request to join the matchmaking queue
type JoinMatchmakingRequest struct {
	League      string `json:"league"`        // ROOKIE, STREET, PRO, TOP_FUEL
	ClientReqID string `json:"client_req_id"` // Client-provided idempotency key (UUID)
}

// Validate validates the request fields
func (r *JoinMatchmakingRequest) Validate() error {
	if r.League == "" {
		return fmt.Errorf("league is required")
	}
	if r.ClientReqID == "" {
		return fmt.Errorf("client_req_id is required")
	}
	
	validLeagues := map[string]bool{
		"ROOKIE": true, "STREET": true, "PRO": true, "TOP_FUEL": true,
	}
	if !validLeagues[r.League] {
		return fmt.Errorf("invalid league: %s", r.League)
	}
	
	return nil
}

// JoinMatchmakingResponse represents a response to joining the matchmaking queue
type JoinMatchmakingResponse struct {
	QueueTicketID        string `json:"queue_ticket_id"`        // Ticket identifier for this queue entry
	Status               string `json:"status"`                 // QUEUED
	EstimatedWaitSeconds int32  `json:"estimated_wait_seconds"` // Estimated wait time (up to 20s)
}

// CancelMatchmakingRequest represents a request to cancel matchmaking
type CancelMatchmakingRequest struct {
	QueueTicketID string `json:"queue_ticket_id"` // Ticket to cancel
	ClientReqID   string `json:"client_req_id"`   // Idempotency key
}

// Validate validates the request fields
func (r *CancelMatchmakingRequest) Validate() error {
	if r.QueueTicketID == "" {
		return fmt.Errorf("queue_ticket_id is required")
	}
	if r.ClientReqID == "" {
		return fmt.Errorf("client_req_id is required")
	}
	return nil
}

// CancelMatchmakingResponse represents a response to canceling matchmaking
type CancelMatchmakingResponse struct {
	Status string `json:"status"` // CANCELLED
}

// Helper functions for JSON marshaling/unmarshaling

// UnmarshalJoinMatchmakingRequest unmarshals JSON data to JoinMatchmakingRequest
func UnmarshalJoinMatchmakingRequest(data []byte) (*JoinMatchmakingRequest, error) {
	var req JoinMatchmakingRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JoinMatchmakingRequest: %w", err)
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid JoinMatchmakingRequest: %w", err)
	}
	return &req, nil
}

// MarshalJoinMatchmakingResponse marshals JoinMatchmakingResponse to JSON
func MarshalJoinMatchmakingResponse(resp *JoinMatchmakingResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// UnmarshalCancelMatchmakingRequest unmarshals JSON data to CancelMatchmakingRequest
func UnmarshalCancelMatchmakingRequest(data []byte) (*CancelMatchmakingRequest, error) {
	var req CancelMatchmakingRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CancelMatchmakingRequest: %w", err)
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid CancelMatchmakingRequest: %w", err)
	}
	return &req, nil
}

// MarshalCancelMatchmakingResponse marshals CancelMatchmakingResponse to JSON
func MarshalCancelMatchmakingResponse(resp *CancelMatchmakingResponse) ([]byte, error) {
	return json.Marshal(resp)
}