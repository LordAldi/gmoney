package pricing

import (
	"errors"
	"math/big"

	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/rate"
)

var ErrInvalidTiers = errors.New("tiers must be defined and sorted")

// Tier defines a range of usage and the price per unit within that range.
type Tier struct {
	UpTo  int64     // The usage limit for this tier (0 means infinite/final tier)
	Price rate.Rate // The price per unit in this tier
}

// CalculateGraduatedCost computes the total cost for a given usage volume.
func CalculateGraduatedCost(usage int64, tiers []Tier) (money.Money, error) {
	if usage < 0 {
		return money.Money{}, errors.New("usage cannot be negative")
	}

	total := new(big.Rat).SetInt64(0)
	remainingUsage := usage

	// 1. Iterate through tiers filling buckets
	for _, tier := range tiers {
		if remainingUsage <= 0 {
			break
		}

		var countInTier int64

		// Check if this is the final tier (UpTo == 0) or if we fit completely
		if tier.UpTo == 0 || remainingUsage <= tier.UpTo {
			// Take everything remaining
			countInTier = remainingUsage
		} else {
			// Fill this bucket and move to next
			countInTier = tier.UpTo
		}

		// Math: cost = count * rate
		tierCost := tier.Price.Mul(countInTier)
		total.Add(total, tierCost)

		// Reduce usage and (crucial) reduce the tier limit logic
		// Actually, standard logic usually defines tiers as "First X", "Next Y".
		// If tiers are defined as "0-1000", "1001-5000", we need to track offsets.
		// For simplicity here, let's assume `UpTo` means "Capacity of this tier".
		// e.g. Tier 1 Capacity: 10,000. Tier 2 Capacity: 40,000.

		remainingUsage -= countInTier
	}

	// 2. Convert high-precision Total to Standard Money (int64 cents)
	// We must define our currency (assuming USD for the rate input for now)

	// FloatString(2) gives us "2400.00" string representation
	// But converting Rat to Int64 requires careful rounding.

	// Convert dollars to cents: Multiply by 100
	centsRat := new(big.Rat).Mul(total, big.NewRat(100, 1))

	// Round Half Up logic on the Big Rat
	// Add 0.5 then floor (standard integer rounding trick)
	half := big.NewRat(1, 2)
	centsRat.Add(centsRat, half)

	// Extract int64
	num := centsRat.Num()
	denom := centsRat.Denom()
	finalCents := new(big.Int).Div(num, denom).Int64()

	return money.New(finalCents, "USD"), nil
}
