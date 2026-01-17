package ledger

import (
	"errors"
	"fmt"
	"time"

	"github.com/LordAldi/gmoney/pkg/money"
)

// AccountType helps us track Assets vs Liabilities
type AccountType int

const (
	Asset     AccountType = iota // Cash, Accounts Receivable
	Liability                    // User Deposits, Deferred Revenue
	Equity                       // Retained Earnings
	Revenue                      // Sales
	Expense                      // COGS, Server Costs
)

type Account struct {
	ID   string
	Type AccountType
	Name string
}

// Entry is one line in a transaction (e.g., "+$100 to Cash")
type Entry struct {
	AccountID string
	Amount    money.Money // Positive = Debit, Negative = Credit (Standard accounting view)
}

// Transaction represents an atomic movement of money.
type Transaction struct {
	ID        string
	PostedAt  time.Time
	Reference string // e.g., "Invoice #123"
	Entries   []Entry
}

// NewTransaction creates a transaction ONLY if it balances.
func NewTransaction(id, ref string, entries []Entry) (Transaction, error) {
	if len(entries) < 2 {
		return Transaction{}, errors.New("transaction must have at least 2 entries")
	}

	// 1. Verify Currency Consistency
	currency := entries[0].Amount.Currency()

	// 2. Verify Balance (Sum of Debits must equal Sum of Credits)
	// In a signed system: Sum must be 0.
	// In a Debit/Credit system: Sum(Debits) == Sum(Credits).
	// Let's use the Signed approach (Debit +, Credit -) for simplicity in storage.

	balance := int64(0)
	for _, e := range entries {
		if e.Amount.Currency() != currency {
			return Transaction{}, fmt.Errorf("mixed currencies in transaction: %s vs %s", currency, e.Amount.Currency())
		}
		balance += e.Amount.Amount()
	}

	if balance != 0 {
		return Transaction{}, fmt.Errorf("transaction does not balance! Diff: %d cents", balance)
	}

	return Transaction{
		ID:        id,
		PostedAt:  time.Now(),
		Reference: ref,
		Entries:   entries,
	}, nil
}
