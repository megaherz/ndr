package constants

import "github.com/shopspring/decimal"

// League name constants
const (
	LeagueRookie  = "ROOKIE"
	LeagueStreet  = "STREET"
	LeaguePro     = "PRO"
	LeagueTopFuel = "TOP_FUEL"
)

// League buy-in amounts
var LeagueBuyins = map[string]decimal.Decimal{
	LeagueRookie:  decimal.NewFromInt(10),   // 10 FUEL
	LeagueStreet:  decimal.NewFromInt(50),   // 50 FUEL
	LeaguePro:     decimal.NewFromInt(300),  // 300 FUEL
	LeagueTopFuel: decimal.NewFromInt(3000), // 3000 FUEL
}

// ValidLeagues returns a slice of all valid league names
func ValidLeagues() []string {
	return []string{
		LeagueRookie,
		LeagueStreet,
		LeaguePro,
		LeagueTopFuel,
	}
}

// IsValidLeague checks if a league name is valid
func IsValidLeague(league string) bool {
	switch league {
	case LeagueRookie, LeagueStreet, LeaguePro, LeagueTopFuel:
		return true
	default:
		return false
	}
}

// GetLeagueBuyin returns the buy-in amount for a league
func GetLeagueBuyin(league string) (decimal.Decimal, bool) {
	buyin, exists := LeagueBuyins[league]
	return buyin, exists
}
