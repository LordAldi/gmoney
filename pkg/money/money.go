// pkg/money/money.go
package money

import (
	"errors"
	"fmt"
)

var (
	ErrMismatchCurrency = errors.New("cannot operate on mismatched currencies")
	ErrInvalidAmount    = errors.New("amount cannot be negative in this context")
)

// Money represents a monetary value in its minor unit (e.g., cents).
// It is immutable by design.
type Money struct {
	amount   int64  // 1000 = $10.00
	currency string // ISO 4217 code "USD", "EUR"
}

func New(amount int64, currency string) Money {
	return Money{amount: amount, currency: currency}
}

func (m Money) Amount() int64    { return m.amount }
func (m Money) Currency() string { return m.currency }

// String implements Stringer for easy debugging
func (m Money) String() string {
	return fmt.Sprintf("%d (%s)", m.amount, m.currency)
}

// Add ensures currency safety
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrMismatchCurrency
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}
