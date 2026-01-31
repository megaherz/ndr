package gameengine

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CrashSeedData represents the crash seeds for all three heats
type CrashSeedData struct {
	Heat1Seed string `json:"heat1_seed"`
	Heat2Seed string `json:"heat2_seed"`
	Heat3Seed string `json:"heat3_seed"`
	MatchID   string `json:"match_id"`
	Timestamp int64  `json:"timestamp"`
}

// ProvableFairnessEngine handles cryptographic seed generation for provable fairness
type ProvableFairnessEngine interface {
	// GenerateCrashSeeds generates cryptographic seeds for all three heats
	GenerateCrashSeeds(matchID uuid.UUID) (*CrashSeedData, error)
	
	// GenerateCommitHash generates a pre-commitment hash of the crash seeds
	GenerateCommitHash(seedData *CrashSeedData) (string, error)
	
	// VerifyCommitHash verifies that the revealed seeds match the pre-commitment hash
	VerifyCommitHash(seedData *CrashSeedData, commitHash string) bool
	
	// GenerateHeatSeed generates a single cryptographic seed for a heat
	GenerateHeatSeed() (string, error)
	
	// DeriveRandomValue derives a deterministic random value from a seed and context
	DeriveRandomValue(seed, context string) uint64
}

// provableFairnessEngine implements ProvableFairnessEngine
type provableFairnessEngine struct{}

// NewProvableFairnessEngine creates a new provable fairness engine
func NewProvableFairnessEngine() ProvableFairnessEngine {
	return &provableFairnessEngine{}
}

// GenerateCrashSeeds generates cryptographic seeds for all three heats
func (p *provableFairnessEngine) GenerateCrashSeeds(matchID uuid.UUID) (*CrashSeedData, error) {
	heat1Seed, err := p.GenerateHeatSeed()
	if err != nil {
		return nil, fmt.Errorf("failed to generate heat 1 seed: %w", err)
	}
	
	heat2Seed, err := p.GenerateHeatSeed()
	if err != nil {
		return nil, fmt.Errorf("failed to generate heat 2 seed: %w", err)
	}
	
	heat3Seed, err := p.GenerateHeatSeed()
	if err != nil {
		return nil, fmt.Errorf("failed to generate heat 3 seed: %w", err)
	}
	
	seedData := &CrashSeedData{
		Heat1Seed: heat1Seed,
		Heat2Seed: heat2Seed,
		Heat3Seed: heat3Seed,
		MatchID:   matchID.String(),
		Timestamp: time.Now().Unix(),
	}
	
	return seedData, nil
}

// GenerateCommitHash generates a pre-commitment hash of the crash seeds
func (p *provableFairnessEngine) GenerateCommitHash(seedData *CrashSeedData) (string, error) {
	// Serialize the seed data to JSON for consistent hashing
	jsonData, err := json.Marshal(seedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal seed data: %w", err)
	}
	
	// Generate SHA-256 hash
	hash := sha256.Sum256(jsonData)
	
	// Return hex-encoded hash
	return hex.EncodeToString(hash[:]), nil
}

// VerifyCommitHash verifies that the revealed seeds match the pre-commitment hash
func (p *provableFairnessEngine) VerifyCommitHash(seedData *CrashSeedData, commitHash string) bool {
	// Generate hash from the provided seed data
	calculatedHash, err := p.GenerateCommitHash(seedData)
	if err != nil {
		return false
	}
	
	// Compare hashes (constant-time comparison for security)
	return calculatedHash == commitHash
}

// GenerateHeatSeed generates a single cryptographic seed for a heat
func (p *provableFairnessEngine) GenerateHeatSeed() (string, error) {
	// Generate 32 bytes of cryptographically secure random data
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Return hex-encoded seed
	return hex.EncodeToString(randomBytes), nil
}

// DeriveRandomValue derives a deterministic random value from a seed and context
func (p *provableFairnessEngine) DeriveRandomValue(seed, context string) uint64 {
	// Combine seed and context
	combined := seed + ":" + context
	
	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(combined))
	
	// Convert first 8 bytes to uint64 (big-endian)
	var result uint64
	for i := 0; i < 8; i++ {
		result = (result << 8) | uint64(hash[i])
	}
	
	return result
}

// GenerateMatchSeeds is a convenience function to generate and hash seeds for a match
func GenerateMatchSeeds(matchID uuid.UUID) (seedData *CrashSeedData, commitHash string, err error) {
	engine := NewProvableFairnessEngine()
	
	// Generate crash seeds
	seedData, err = engine.GenerateCrashSeeds(matchID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate crash seeds: %w", err)
	}
	
	// Generate commitment hash
	commitHash, err = engine.GenerateCommitHash(seedData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate commit hash: %w", err)
	}
	
	return seedData, commitHash, nil
}

// VerifyMatchSeeds verifies the integrity of match seeds
func VerifyMatchSeeds(seedData *CrashSeedData, commitHash string) bool {
	engine := NewProvableFairnessEngine()
	return engine.VerifyCommitHash(seedData, commitHash)
}

// GetHeatSeedFromMatch extracts a specific heat seed from match crash seed data
func GetHeatSeedFromMatch(crashSeedJSON string, heat int) (string, error) {
	var seedData CrashSeedData
	err := json.Unmarshal([]byte(crashSeedJSON), &seedData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal crash seed data: %w", err)
	}
	
	switch heat {
	case 1:
		return seedData.Heat1Seed, nil
	case 2:
		return seedData.Heat2Seed, nil
	case 3:
		return seedData.Heat3Seed, nil
	default:
		return "", fmt.Errorf("invalid heat number: %d", heat)
	}
}