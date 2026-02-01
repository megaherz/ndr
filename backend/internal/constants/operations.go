package constants

// Ledger operation type constants
const (
	OperationDeposit         = "DEPOSIT"
	OperationWithdrawal      = "WITHDRAWAL"
	OperationMatchBuyin      = "MATCH_BUYIN"
	OperationMatchPrize      = "MATCH_PRIZE"
	OperationMatchRake       = "MATCH_RAKE"
	OperationMatchBurnReward = "MATCH_BURN_REWARD"
	OperationInitialBalance  = "INITIAL_BALANCE"
)

// ValidOperationTypes returns a slice of all valid operation types
func ValidOperationTypes() []string {
	return []string{
		OperationDeposit,
		OperationWithdrawal,
		OperationMatchBuyin,
		OperationMatchPrize,
		OperationMatchRake,
		OperationMatchBurnReward,
		OperationInitialBalance,
	}
}

// IsValidOperationType checks if an operation type string is valid
func IsValidOperationType(operationType string) bool {
	switch operationType {
	case OperationDeposit, OperationWithdrawal, OperationMatchBuyin,
		OperationMatchPrize, OperationMatchRake, OperationMatchBurnReward,
		OperationInitialBalance:
		return true
	default:
		return false
	}
}
