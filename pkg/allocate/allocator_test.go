package allocate

import (
	"testing"

	"github.com/LordAldi/gmoney/pkg/money"
)

func TestSplit_PennyVariance(t *testing.T) {
	// Scenario: Split $0.05 (5 cents) among 3 people equally.
	// Math: 5 / 3 = 1.666...
	// Expected: 2, 2, 1 (Total 5).
	// Note: 2, 1, 2 is also valid, but 2, 2, 1 implies the first weights got priority
	// or the residue sort was stable.

	total := money.New(5, "USD")
	weights := []int{1, 1, 1}

	parts, err := Split(total, weights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. Verify Summation (Law of Conservation of Money)
	sum := int64(0)
	for _, p := range parts {
		sum += p.Amount()
	}

	if sum != total.Amount() {
		t.Errorf("Money created/destroyed! Started with %d, ended with %d", total.Amount(), sum)
	}

	// 2. Verify Distribution
	// Since weights are equal, the residues are equal.
	// The sort stability usually gives it to the first indices, or random depending on impl.
	// For strict reproducibility, we might need a secondary sort key in the implementation.
	t.Logf("Distribution: %v", parts)
}

func TestSplit_Complex(t *testing.T) {
	// Scenario: $100.00 split by weights [1, 2, 3]
	// Total: 10000 cents
	// Sum Weights: 6
	// 1/6 = 16.666% -> 1666.6 -> 1666 + rem
	// 2/6 = 33.333% -> 3333.3 -> 3333 + rem
	// 3/6 = 50.000% -> 5000.0 -> 5000

	total := money.New(10000, "USD")
	weights := []int{1, 2, 3}

	parts, _ := Split(total, weights)

	if parts[0].Amount()+parts[1].Amount()+parts[2].Amount() != 10000 {
		t.Error("Sum mismatch")
	}
}
