package constants

// Currency constants for the game economy
const (
	CurrencyTON  = "TON"
	CurrencyFUEL = "FUEL"
	CurrencyBURN = "BURN"
)

// ValidCurrencies returns a slice of all valid currency types
func ValidCurrencies() []string {
	return []string{CurrencyTON, CurrencyFUEL, CurrencyBURN}
}

// IsValidCurrency checks if a currency string is valid
func IsValidCurrency(currency string) bool {
	switch currency {
	case CurrencyTON, CurrencyFUEL, CurrencyBURN:
		return true
	default:
		return false
	}
}
