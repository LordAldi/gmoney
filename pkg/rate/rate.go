package rate

import (
	"fmt"
	"math/big"
)

// Rate represents a unit price with high precision (e.g., $0.0045).
// It wraps big.Rat to handle fractions perfectly.
type Rate struct {
	value *big.Rat
}

// New creates a rate from a string to avoid float imprecision.
// Example: New("0.0045")
func New(amount string) (Rate, error) {
	r, ok := new(big.Rat).SetString(amount)
	if !ok {
		return Rate{}, fmt.Errorf("invalid rate format: %s", amount)
	}
	return Rate{value: r}, nil
}

// Mul multiplies the Rate by a quantity (int64) and returns a big.Rat result.
// We DO NOT return money.Money yet, because we might need to sum multiple
// fractional tiers before rounding.
func (r Rate) Mul(quantity int64) *big.Rat {
	q := new(big.Rat).SetInt64(quantity)
	return new(big.Rat).Mul(r.value, q)
}

// Raw returns the underlying big.Rat value of the Rate.
func (r Rate) Raw() *big.Rat {
	return r.value
}
