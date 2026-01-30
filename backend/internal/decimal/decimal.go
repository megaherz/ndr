package decimal

import (
	"database/sql/driver"
	"fmt"

	"github.com/shopspring/decimal"
)

// All monetary values in the system use fixed-point decimal arithmetic
// with 2 decimal places and ALWAYS round down (floor) to prevent
// accumulation of rounding errors in the economy.

// Zero represents decimal zero
var Zero = decimal.Zero

// NewFromString creates a decimal from string with validation
func NewFromString(value string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid decimal string: %s", value)
	}
	return d, nil
}

// NewFromFloat64 creates a decimal from float64 (use with caution)
func NewFromFloat64(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}

// NewFromInt creates a decimal from int64
func NewFromInt(value int64) decimal.Decimal {
	return decimal.NewFromInt(value)
}

// MustFromString creates a decimal from string and panics on error
func MustFromString(value string) decimal.Decimal {
	d, err := NewFromString(value)
	if err != nil {
		panic(err)
	}
	return d
}

// ToMonetary converts a decimal to monetary format (2 decimal places, rounded down)
func ToMonetary(d decimal.Decimal) decimal.Decimal {
	// Always round down (floor) to 2 decimal places
	return d.Truncate(2)
}

// ToMonetaryString converts a decimal to monetary string format
func ToMonetaryString(d decimal.Decimal) string {
	return ToMonetary(d).StringFixed(2)
}

// Add performs monetary addition (result rounded down to 2 decimal places)
func Add(a, b decimal.Decimal) decimal.Decimal {
	return ToMonetary(a.Add(b))
}

// Sub performs monetary subtraction (result rounded down to 2 decimal places)
func Sub(a, b decimal.Decimal) decimal.Decimal {
	return ToMonetary(a.Sub(b))
}

// Mul performs monetary multiplication (result rounded down to 2 decimal places)
func Mul(a, b decimal.Decimal) decimal.Decimal {
	return ToMonetary(a.Mul(b))
}

// Div performs monetary division (result rounded down to 2 decimal places)
func Div(a, b decimal.Decimal) decimal.Decimal {
	if b.IsZero() {
		panic("division by zero")
	}
	return ToMonetary(a.Div(b))
}

// DivRoundDown performs division with explicit round down
func DivRoundDown(a, b decimal.Decimal) decimal.Decimal {
	if b.IsZero() {
		panic("division by zero")
	}
	return a.Div(b).Truncate(2)
}

// Percentage calculates percentage of a value (rounded down)
func Percentage(value decimal.Decimal, percent decimal.Decimal) decimal.Decimal {
	return ToMonetary(value.Mul(percent).Div(decimal.NewFromInt(100)))
}

// IsPositive returns true if decimal is greater than zero
func IsPositive(d decimal.Decimal) bool {
	return d.GreaterThan(decimal.Zero)
}

// IsNegative returns true if decimal is less than zero
func IsNegative(d decimal.Decimal) bool {
	return d.LessThan(decimal.Zero)
}

// IsZero returns true if decimal equals zero
func IsZero(d decimal.Decimal) bool {
	return d.Equal(decimal.Zero)
}

// Min returns the smaller of two decimals
func Min(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

// Max returns the larger of two decimals
func Max(a, b decimal.Decimal) decimal.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}

// Abs returns the absolute value of a decimal
func Abs(d decimal.Decimal) decimal.Decimal {
	return d.Abs()
}

// SumMonetary sums a slice of decimals with monetary rounding
func SumMonetary(values []decimal.Decimal) decimal.Decimal {
	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	return ToMonetary(sum)
}

// League buy-in amounts (constants)
var (
	RookieBuyIn  = MustFromString("10.00")   // 10 FUEL
	StreetBuyIn  = MustFromString("50.00")   // 50 FUEL
	ProBuyIn     = MustFromString("300.00")  // 300 FUEL
	TopFuelBuyIn = MustFromString("3000.00") // 3000 FUEL
)

// GetLeagueBuyIn returns the buy-in amount for a league
func GetLeagueBuyIn(league string) (decimal.Decimal, error) {
	switch league {
	case "ROOKIE":
		return RookieBuyIn, nil
	case "STREET":
		return StreetBuyIn, nil
	case "PRO":
		return ProBuyIn, nil
	case "TOP_FUEL":
		return TopFuelBuyIn, nil
	default:
		return decimal.Zero, fmt.Errorf("unknown league: %s", league)
	}
}

// Rake percentage (8%)
var RakePercentage = MustFromString("8.00")

// CalculateRake calculates 8% rake from total buy-ins
func CalculateRake(totalBuyins decimal.Decimal) decimal.Decimal {
	return Percentage(totalBuyins, RakePercentage)
}

// CalculatePrizePool calculates prize pool after rake deduction
func CalculatePrizePool(totalBuyins decimal.Decimal) decimal.Decimal {
	rake := CalculateRake(totalBuyins)
	return Sub(totalBuyins, rake)
}

// Prize distribution percentages (after rake)
var (
	FirstPlacePct  = MustFromString("50.00") // 50% of prize pool
	SecondPlacePct = MustFromString("30.00") // 30% of prize pool
	ThirdPlacePct  = MustFromString("20.00") // 20% of prize pool
)

// CalculatePrizes calculates prize distribution for top 3 positions
func CalculatePrizes(prizePool decimal.Decimal) (first, second, third decimal.Decimal) {
	first = Percentage(prizePool, FirstPlacePct)
	second = Percentage(prizePool, SecondPlacePct)
	third = Percentage(prizePool, ThirdPlacePct)
	return first, second, third
}

// ValidateMonetary validates that a decimal is a valid monetary amount
func ValidateMonetary(d decimal.Decimal) error {
	if d.IsNegative() {
		return fmt.Errorf("monetary amount cannot be negative: %s", d.String())
	}

	// Check that it has at most 2 decimal places
	if d.Exponent() < -2 {
		return fmt.Errorf("monetary amount cannot have more than 2 decimal places: %s", d.String())
	}

	return nil
}

// NullDecimal represents a nullable decimal for database operations
type NullDecimal struct {
	Decimal decimal.Decimal
	Valid   bool
}

// Scan implements the sql.Scanner interface
func (nd *NullDecimal) Scan(value interface{}) error {
	if value == nil {
		nd.Decimal, nd.Valid = decimal.Zero, false
		return nil
	}

	nd.Valid = true
	return nd.Decimal.Scan(value)
}

// Value implements the driver.Valuer interface
func (nd NullDecimal) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	}
	return nd.Decimal.Value()
}

// String returns string representation
func (nd NullDecimal) String() string {
	if !nd.Valid {
		return "NULL"
	}
	return nd.Decimal.String()
}

// NewNullDecimal creates a new valid NullDecimal
func NewNullDecimal(d decimal.Decimal) NullDecimal {
	return NullDecimal{
		Decimal: d,
		Valid:   true,
	}
}

// NewNullDecimalFromString creates a NullDecimal from string
func NewNullDecimalFromString(s string) (NullDecimal, error) {
	if s == "" {
		return NullDecimal{Valid: false}, nil
	}

	d, err := NewFromString(s)
	if err != nil {
		return NullDecimal{Valid: false}, err
	}

	return NewNullDecimal(d), nil
}
