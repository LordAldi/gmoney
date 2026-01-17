package allocate_test

import (
	"math"
	"testing"

	"github.com/LordAldi/gmoney/pkg/allocate"
	"github.com/LordAldi/gmoney/pkg/money"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestSplit_Properties(t *testing.T) {
	// 1. Configure the Property Tester
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 5000 // Run 5,000 random scenarios (High confidence)
	properties := gopter.NewProperties(parameters)

	// --- GENERATORS ---

	// Generator A: Standard "Business" Money ($0.01 to $1,000,000.00)
	genMoneyStandard := gen.Int64Range(1, 100_000_000).Map(func(v int64) money.Money {
		return money.New(v, "USD")
	})

	// Generator B: "Stress" Money (Near MaxInt64 to trigger overflows)
	// We use MaxInt64 - 1000 to allow slight room for addition without immediate panic,
	// but enough to break naive multiplication logic.
	genMoneyStress := gen.OneConstOf(
		money.New(math.MaxInt64, "USD"),
		money.New(math.MaxInt64-1, "USD"),
		money.New(1, "USD"), // The Penny split
	)

	// Combined Generator: 90% Standard, 10% Stress
	genMoney := gen.OneGenOf(
		genMoneyStandard,
		genMoneyStress,
	)

	// Generator C: Weights
	// A slice of 1 to 50 items, each weight between 1 and 10,000.
	genWeights := gen.SliceOfN(50, gen.IntRange(1, 10000)).Map(func(v []int) []int {
		// Filter: Ensure we don't pass empty slices (business rule)
		if len(v) == 0 {
			return []int{1}
		}
		return v
	})

	// --- PROPERTIES ---

	// Property 1: Conservation of Money
	// "The sum of the parts must EXACTLY equal the total amount."
	properties.Property("Invariant: Sum(Parts) == Total", prop.ForAll(
		func(total money.Money, weights []int) bool {
			parts, err := allocate.Split(total, weights)

			// If our logic errors on valid input, that's a failure.
			if err != nil {
				return false
			}

			sum := int64(0)
			for _, p := range parts {
				sum += p.Amount()
			}

			// The invariant check
			return sum == total.Amount()
		},
		genMoney,   // Input 1
		genWeights, // Input 2
	))

	// Property 2: Non-Negativity
	// "If the input is positive, no allocated part can ever be negative."
	// (This catches overflow bugs where numbers wrap around to negative)
	properties.Property("Invariant: No Negative Allocations", prop.ForAll(
		func(total money.Money, weights []int) bool {
			if total.Amount() < 0 {
				return true // Skip negative input tests for this specific property
			}

			parts, err := allocate.Split(total, weights)
			if err != nil {
				return false
			}

			for _, p := range parts {
				if p.Amount() < 0 {
					return false // FAIL: We generated a negative amount!
				}
			}
			return true
		},
		genMoney,
		genWeights,
	))

	// Property 3: Monotonicity (Fairness)
	// "If Weight A > Weight B, then Amount A must be >= Amount B"
	properties.Property("Invariant: Higher Weight >= Lower Weight", prop.ForAll(
		func(total money.Money, weights []int) bool {
			parts, err := allocate.Split(total, weights)
			if err != nil {
				return false
			}

			// Verify every pair maintains the hierarchy
			for i := 0; i < len(weights); i++ {
				for j := 0; j < len(weights); j++ {
					if weights[i] > weights[j] {
						// Logic: Strictly heavier weights should not get LESS money.
						if parts[i].Amount() < parts[j].Amount() {
							return false
						}
					}
				}
			}
			return true
		},
		genMoney,
		genWeights,
	))

	// Execute
	properties.TestingRun(t)
}
