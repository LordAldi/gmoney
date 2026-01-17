package examples_test

import (
	"fmt"
	"testing"

	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/refund"
)

func TestRefundScenarios(t *testing.T) {
	// ==========================================
	// SETUP: The Original Order
	// ==========================================
	// Item: High-End Monitor
	// Qty: 3
	// Unit Price: $1,000.00
	// Tax Rate: 25% (Exclusive)
	//
	// Total Base: $3,000.00 (3 * 1000)
	// Total Tax:    $750.00 (3000 * 0.25)
	// Grand Total: $3,750.00

	originalOrder := []refund.Component{
		{Name: "Base Price", Amount: money.New(300000, "USD")}, // $3000.00
		{Name: "VAT (25%) ", Amount: money.New(75000, "USD")},  // $750.00
	}

	// Helper to calculate percent of total (for negotiation cases)
	totalOrderAmount := int64(375000) // $3,750.00
	calcPercent := func(percent float64) money.Money {
		val := float64(totalOrderAmount) * percent
		return money.New(int64(val), "USD")
	}

	fmt.Println("--- REFUND SCENARIOS LOG ---")
	fmt.Printf("Original Order Total: $3,750.00 (Base: $3,000 | Tax: $750)\n\n")

	// ==========================================
	// CASE 1: Full Refund (3 Qty -> 3 Returned)
	// ==========================================
	// Logic: Refund 100% of the money.
	t.Run("1. Full Refund", func(t *testing.T) {
		refundAmount := money.New(375000, "USD") // Full $3,750.00

		res, _ := refund.CalculateNegotiatedRefund(originalOrder, refundAmount)

		printResult("1. Full Refund (3/3)", res)
	})

	// ==========================================
	// CASE 2: Partial Quantity (3 Qty -> 1 Returned)
	// ==========================================
	// Logic: Refund exactly 1/3 of the total value.
	// Math: $3,750 / 3 = $1,250.00
	t.Run("2. Partial Quantity", func(t *testing.T) {
		refundAmount := money.New(totalOrderAmount/3, "USD") // $1,250.00

		res, _ := refund.CalculateNegotiatedRefund(originalOrder, refundAmount)

		printResult("2. Return 1 Item (1/3)", res)
	})

	// ==========================================
	// CASE 3: Negotiated 30% (Defective but Kept)
	// ==========================================
	// Logic: Customer keeps all 3, but gets 30% money back.
	// Refund Total: $3,750 * 0.30 = $1,125.00
	t.Run("3. Negotiated 30%", func(t *testing.T) {
		refundAmount := calcPercent(0.30) // $1,125.00

		res, _ := refund.CalculateNegotiatedRefund(originalOrder, refundAmount)

		printResult("3. Negotiated (30%)", res)
	})

	// ==========================================
	// CASE 4: Negotiated 67% (Major Damage)
	// ==========================================
	// Logic: Refund 67% of total.
	// Refund Total: $3,750 * 0.67 = $2,512.50
	t.Run("4. Negotiated 67%", func(t *testing.T) {
		refundAmount := calcPercent(0.67) // $2,512.50

		res, _ := refund.CalculateNegotiatedRefund(originalOrder, refundAmount)

		printResult("4. Negotiated (67%)", res)
	})

	// ==========================================
	// CASE 5: The "Fixed Amount" Negotiation
	// ==========================================
	// Logic: "Just give me $666 back and we are good."
	// Challenge: Split $666.00 across Base ($3000) and Tax ($750).
	// Ratio: 80% Base / 20% Tax.
	t.Run("5. Fixed Amount $666", func(t *testing.T) {
		refundAmount := money.New(66600, "USD") // $666.00

		res, _ := refund.CalculateNegotiatedRefund(originalOrder, refundAmount)

		printResult("5. Negotiated ($666)", res)
	})
}

// Helper to format output neatly
func printResult(title string, res refund.RefundResult) {
	fmt.Println(title)
	fmt.Printf("   Refund Total: %s\n", res.Total)
	for _, c := range res.Components {
		fmt.Printf("   - %s: %s\n", c.Name, c.Amount)
	}
	fmt.Println("")
}
