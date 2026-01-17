package refund

import (
	"errors"

	"github.com/LordAldi/gmoney/pkg/allocate"
	"github.com/LordAldi/gmoney/pkg/money"
)

// Component represents a line on the original invoice (e.g., "Base Price", "VAT").
type Component struct {
	Name   string
	Amount money.Money
}

// RefundResult tells you exactly how much to refund per component.
type RefundResult struct {
	Components []Component
	Total      money.Money
}

// CalculateNegotiatedRefund distributes a flat refund amount across the original components
// to ensure the refund maintains the exact same tax/base ratio as the original order.
func CalculateNegotiatedRefund(originalParts []Component, refundAmount money.Money) (RefundResult, error) {
	// 1. Validate: Cannot refund more than the original total
	originalTotalAmt := int64(0)
	weights := make([]int, len(originalParts))

	for i, part := range originalParts {
		// The Weight is the amount of the original component.
		// If Base was $1000 and Tax was $250, weights are [1000, 250].
		weights[i] = int(part.Amount.Amount())
		originalTotalAmt += part.Amount.Amount()
	}

	if refundAmount.Amount() > originalTotalAmt {
		return RefundResult{}, errors.New("refund amount exceeds original order total")
	}

	// 2. The Magic: Use your existing Allocator
	// We split the "Refund Amount" using the "Original Amounts" as weights.
	allocatedRefunds, err := allocate.Split(refundAmount, weights)
	if err != nil {
		return RefundResult{}, err
	}

	// 3. Reconstruct Result
	result := RefundResult{
		Total:      refundAmount,
		Components: make([]Component, len(originalParts)),
	}

	for i, refAmt := range allocatedRefunds {
		result.Components[i] = Component{
			Name:   originalParts[i].Name,
			Amount: refAmt,
		}
	}

	return result, nil
}
