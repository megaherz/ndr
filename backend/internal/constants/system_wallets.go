package constants

// System wallet name constants
const (
	SystemWalletHouseFuel = "HOUSE_FUEL"
	SystemWalletRakeFuel  = "RAKE_FUEL"
)

// ValidSystemWallets returns a slice of all valid system wallet names
func ValidSystemWallets() []string {
	return []string{
		SystemWalletHouseFuel,
		SystemWalletRakeFuel,
	}
}

// IsValidSystemWallet checks if a system wallet name is valid
func IsValidSystemWallet(walletName string) bool {
	switch walletName {
	case SystemWalletHouseFuel, SystemWalletRakeFuel:
		return true
	default:
		return false
	}
}
