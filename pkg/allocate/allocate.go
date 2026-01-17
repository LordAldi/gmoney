package allocate

import (
	"errors"
	"math/big"
	"sort"

	"github.com/LordAldi/gmoney/pkg/money"
)

var ErrInvalidWeights = errors.New("weights must be greater than zero")

type SplitResult struct {
	Original money.Money
	Parts    []money.Money
}

// Split allocates a total amount into parts based on the provided weights.
// It uses the Largest Remainder Method to ensure accuracy.
func Split(total money.Money, weights []int) ([]money.Money, error) {
	if len(weights) == 0 {
		return nil, ErrInvalidWeights
	}

	totalAmount := total.Amount()
	sumWeights := int64(0)
	for _, w := range weights {
		sumWeights += int64(w)
	}

	if sumWeights == 0 {
		return nil, ErrInvalidWeights
	}

	results := make([]money.Money, len(weights))
	remainder := totalAmount

	// Structure to track which part deserves the "remainder penny" most
	type entry struct {
		index   int
		base    int64 // The floored amount
		residue int64 // The remainder of the division (the claim to the penny)
	}

	entries := make([]entry, len(weights))

	// Pass 1: Calculate Base Shares and Residues
	for i, w := range weights {
		// Use integer math: (Total * Weight) / SumWeights
		// We use big logic here conceptually, but int64 fits for standard money
		// Note: A real prod app needs overflow protection here if Total * Weight > 2^63
		w64 := int64(w)

		// New (Safe):
		bTotal := big.NewInt(totalAmount)
		bWeight := big.NewInt(w64)
		bSum := big.NewInt(sumWeights)

		// (Total * Weight)
		bBase := new(big.Int).Mul(bTotal, bWeight)

		// (Total * Weight) / Sum
		bBase.Div(bBase, bSum)

		// Cast back to int64 (safe because result must be <= Total, which fits in int64)
		base := bBase.Int64()

		// Calculate residue: (Total * Weight) % Sum
		bResidue := new(big.Int).Mul(bTotal, bWeight)
		bResidue.Mod(bResidue, bSum)
		residue := bResidue.Int64()

		entries[i] = entry{
			index:   i,
			base:    base,
			residue: residue,
		}

		remainder -= base
	}

	// Pass 2: Distribute the Remainder
	// We sort by Residue (descending). The parts with the highest
	// truncation loss get the remaining pennies.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].residue > entries[j].residue
	})

	// Add 1 unit to the top candidates until remainder is 0
	for i := 0; i < int(remainder); i++ {
		entries[i].base++
	}

	// Pass 3: Reconstruct the results in original order
	// We need to map back because sorting messed up the order
	finalParts := make([]int64, len(weights))
	for _, e := range entries {
		finalParts[e.index] = e.base
	}

	// Convert back to Money objects
	for i, amt := range finalParts {
		results[i] = money.New(amt, total.Currency())
	}

	return results, nil
}
