package refund

import (
	"errors"
	"fmt"

	"github.com/LordAldi/gmoney/pkg/allocate"
	"github.com/LordAldi/gmoney/pkg/money"
)

// LineItem represents a full order line with quantity.
type LineItem struct {
	Quantity   int64       // e.g., 3
	Components []Component // Totals for the WHOLE line (e.g., Base $3000, Tax $750)
}

// RefundCalculation holds the result of the logic.
type RefundCalculation struct {
	MaxRefundable money.Money // The Cap (Value of the returned Qty)
	RefundedParts []Component // The actual allocated refund breakdown
}

// CalculateItemizedRefund handles the 2-level logic:
// 1. Calculate the MAX value for the specific 'returnQty'.
// 2. Allocates the 'negotiatedAmount' across that specific quantity's components.
func CalculateItemizedRefund(line LineItem, returnQty int64, negotiatedAmount money.Money) (RefundCalculation, error) {
	if returnQty > line.Quantity {
		return RefundCalculation{}, errors.New("cannot return more items than purchased")
	}
	if returnQty <= 0 {
		return RefundCalculation{}, errors.New("return quantity must be positive")
	}

	// --- LEVEL 1: Quantity Proration (Calculate The Cap) ---

	// We need to find the value of the SPECIFIC items being returned.
	// Formula: (ComponentTotal / TotalQty) * ReturnQty

	scopeWeights := make([]int, len(line.Components))
	scopeComponents := make([]Component, len(line.Components))

	maxRefundableTotal := int64(0)

	for i, comp := range line.Components {
		// 1. Get Unit Value (Integer math safe? Yes, if we assume money is atomic.
		// For strict precision, we might need allocate.Split here too,
		// but simple division usually suffices for determining the "Cap").

		// Let's use allocate.Split to be 100% safe against "3 items for $10.00" scenarios.
		// We split the Component Amount into [ReturnQty, KeepQty]
		splitParts, _ := allocate.Split(comp.Amount, []int{int(returnQty), int(line.Quantity - returnQty)})

		valOfReturnedItems := splitParts[0] // This is the value of the items we are touching

		scopeComponents[i] = Component{Name: comp.Name, Amount: valOfReturnedItems}
		scopeWeights[i] = int(valOfReturnedItems.Amount())
		maxRefundableTotal += valOfReturnedItems.Amount()
	}

	maxRefundable := money.New(maxRefundableTotal, negotiatedAmount.Currency())

	// --- VALIDATION: The Negotiation Limit ---
	if negotiatedAmount.Amount() > maxRefundableTotal {
		return RefundCalculation{}, fmt.Errorf(
			"negotiated amount %s exceeds value of returned items %s",
			negotiatedAmount, maxRefundable,
		)
	}

	// --- LEVEL 2: Negotiated Allocation ---

	// We distribute the Negotiated Amount based on the weights of the RETURNED items only.
	// If Negotiated == Max, this returns the full scope.
	// If Negotiated < Max, this prorates proportionally.

	allocatedAmounts, err := allocate.Split(negotiatedAmount, scopeWeights)
	if err != nil {
		return RefundCalculation{}, err
	}

	// Reconstruct the breakdown
	finalParts := make([]Component, len(line.Components))
	for i, amt := range allocatedAmounts {
		finalParts[i] = Component{
			Name:   line.Components[i].Name,
			Amount: amt,
		}
	}

	return RefundCalculation{
		MaxRefundable: maxRefundable,
		RefundedParts: finalParts,
	}, nil
}
