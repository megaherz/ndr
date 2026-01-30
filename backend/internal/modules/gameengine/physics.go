package gameengine

import (
	"math"

	"github.com/shopspring/decimal"
)

const (
	// MaxHeatDuration is the maximum duration of a heat in seconds
	MaxHeatDuration = 25.0
	
	// MaxSpeed is the maximum speed achievable at t=25 seconds
	MaxSpeed = 500.0
	
	// SpeedGrowthRate is the exponential growth rate (0.08)
	SpeedGrowthRate = 0.08
)

// PhysicsEngine handles all game physics calculations
type PhysicsEngine interface {
	// CalculateSpeed calculates the speed at time t using the exponential formula
	// Speed = 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1))
	CalculateSpeed(timeSeconds float64) decimal.Decimal
	
	// CalculateTimeForSpeed calculates the time needed to reach a specific speed
	// This is the inverse of CalculateSpeed
	CalculateTimeForSpeed(speed decimal.Decimal) float64
	
	// IsValidSpeed checks if a speed value is achievable within the heat duration
	IsValidSpeed(speed decimal.Decimal) bool
	
	// GetMaxSpeedAtTime returns the maximum possible speed at a given time
	GetMaxSpeedAtTime(timeSeconds float64) decimal.Decimal
}

// physicsEngine implements PhysicsEngine
type physicsEngine struct{}

// NewPhysicsEngine creates a new physics engine
func NewPhysicsEngine() PhysicsEngine {
	return &physicsEngine{}
}

// CalculateSpeed calculates the speed at time t using the exponential formula
// Speed = 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1))
func (p *physicsEngine) CalculateSpeed(timeSeconds float64) decimal.Decimal {
	// Clamp time to valid range
	if timeSeconds < 0 {
		timeSeconds = 0
	}
	if timeSeconds > MaxHeatDuration {
		timeSeconds = MaxHeatDuration
	}
	
	// Calculate e^(0.08 * t)
	expTerm := math.Exp(SpeedGrowthRate * timeSeconds)
	
	// Calculate e^(0.08 * 25) for the denominator
	expMax := math.Exp(SpeedGrowthRate * MaxHeatDuration)
	
	// Apply the formula: 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1))
	speed := MaxSpeed * ((expTerm - 1) / (expMax - 1))
	
	// Convert to decimal with 2 decimal places (truncated, not rounded)
	speedDecimal := decimal.NewFromFloat(speed).Truncate(2)
	
	return speedDecimal
}

// CalculateTimeForSpeed calculates the time needed to reach a specific speed
// This is the inverse of CalculateSpeed
func (p *physicsEngine) CalculateTimeForSpeed(speed decimal.Decimal) float64 {
	speedFloat, _ := speed.Float64()
	
	// Handle edge cases
	if speedFloat <= 0 {
		return 0
	}
	if speedFloat >= MaxSpeed {
		return MaxHeatDuration
	}
	
	// Calculate e^(0.08 * 25) for the denominator
	expMax := math.Exp(SpeedGrowthRate * MaxHeatDuration)
	
	// Solve for t: speed = 500 * ((e^(0.08·t) - 1) / (e^(0.08·25) - 1))
	// Rearranging: e^(0.08·t) = 1 + (speed * (e^(0.08·25) - 1)) / 500
	expTerm := 1 + (speedFloat*(expMax-1))/MaxSpeed
	
	// t = ln(expTerm) / 0.08
	timeSeconds := math.Log(expTerm) / SpeedGrowthRate
	
	// Clamp to valid range
	if timeSeconds < 0 {
		timeSeconds = 0
	}
	if timeSeconds > MaxHeatDuration {
		timeSeconds = MaxHeatDuration
	}
	
	return timeSeconds
}

// IsValidSpeed checks if a speed value is achievable within the heat duration
func (p *physicsEngine) IsValidSpeed(speed decimal.Decimal) bool {
	speedFloat, _ := speed.Float64()
	return speedFloat >= 0 && speedFloat <= MaxSpeed
}

// GetMaxSpeedAtTime returns the maximum possible speed at a given time
func (p *physicsEngine) GetMaxSpeedAtTime(timeSeconds float64) decimal.Decimal {
	return p.CalculateSpeed(timeSeconds)
}

// ValidateSpeedForTime validates that a locked speed is achievable at the given time
func ValidateSpeedForTime(speed decimal.Decimal, timeSeconds float64) bool {
	physics := NewPhysicsEngine()
	maxPossibleSpeed := physics.GetMaxSpeedAtTime(timeSeconds)
	return speed.LessThanOrEqual(maxPossibleSpeed)
}

// GetSpeedAtPercentage calculates speed at a percentage of max heat duration
// Useful for testing and ghost replay generation
func GetSpeedAtPercentage(percentage float64) decimal.Decimal {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	
	timeSeconds := (percentage / 100.0) * MaxHeatDuration
	physics := NewPhysicsEngine()
	return physics.CalculateSpeed(timeSeconds)
}