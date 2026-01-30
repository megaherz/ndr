// Code generated manually from match.proto. DO NOT EDIT.
// Source: match.proto

package match

import (
	"encoding/json"
	"fmt"
)

// EarnPointsRequest represents a request to lock score for a heat
type EarnPointsRequest struct {
	MatchID     string `json:"match_id"`      // Match identifier
	HeatNumber  int32  `json:"heat_number"`   // Heat number (1, 2, or 3)
	ClientReqID string `json:"client_req_id"` // Idempotency key
}

// Validate validates the request fields
func (r *EarnPointsRequest) Validate() error {
	if r.MatchID == "" {
		return fmt.Errorf("match_id is required")
	}
	if r.HeatNumber < 1 || r.HeatNumber > 3 {
		return fmt.Errorf("heat_number must be 1, 2, or 3")
	}
	if r.ClientReqID == "" {
		return fmt.Errorf("client_req_id is required")
	}
	return nil
}

// EarnPointsResponse represents a response to locking score
type EarnPointsResponse struct {
	HeatScore       string `json:"heat_score"`       // Heat score locked (decimal string, e.g., "245.50")
	TotalScore      string `json:"total_score"`      // Total score across all heats (decimal string)
	CurrentPosition int32  `json:"current_position"` // Current position (1-10, may change)
}

// GiveUpRequest represents a request to voluntarily exit a match
type GiveUpRequest struct {
	MatchID     string `json:"match_id"`      // Match identifier
	ClientReqID string `json:"client_req_id"` // Idempotency key
}

// Validate validates the request fields
func (r *GiveUpRequest) Validate() error {
	if r.MatchID == "" {
		return fmt.Errorf("match_id is required")
	}
	if r.ClientReqID == "" {
		return fmt.Errorf("client_req_id is required")
	}
	return nil
}

// GiveUpResponse represents a response to giving up
type GiveUpResponse struct {
	Status      string `json:"status"`       // GAVE_UP
	CurrentHeat string `json:"current_heat"` // Heat number where give up occurred
}

// Helper functions for JSON marshaling/unmarshaling

// UnmarshalEarnPointsRequest unmarshals JSON data to EarnPointsRequest
func UnmarshalEarnPointsRequest(data []byte) (*EarnPointsRequest, error) {
	var req EarnPointsRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EarnPointsRequest: %w", err)
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid EarnPointsRequest: %w", err)
	}
	return &req, nil
}

// MarshalEarnPointsResponse marshals EarnPointsResponse to JSON
func MarshalEarnPointsResponse(resp *EarnPointsResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// UnmarshalGiveUpRequest unmarshals JSON data to GiveUpRequest
func UnmarshalGiveUpRequest(data []byte) (*GiveUpRequest, error) {
	var req GiveUpRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GiveUpRequest: %w", err)
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid GiveUpRequest: %w", err)
	}
	return &req, nil
}

// MarshalGiveUpResponse marshals GiveUpResponse to JSON
func MarshalGiveUpResponse(resp *GiveUpResponse) ([]byte, error) {
	return json.Marshal(resp)
}
