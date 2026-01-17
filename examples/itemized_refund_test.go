package examples_test

import (
	"fmt"
	"testing"

	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/refund"
)

func TestRefund_TwoLevelProration(t *testing.T) {
	// SETUP:
	// Order: 3 Qty.
	// Total Price: $3,750.00.
	// Breakdown: Base $3,000.00 | Tax $750.00.
	line := refund.LineItem{
		Quantity: 3,
		Components: []refund.Component{
			{Name: "Base", Amount: money.New(300000, "USD")},
			{Name: "Tax", Amount: money.New(75000, "USD")},
		},
	}

	fmt.Println("--- TWO-LEVEL REFUND LOG ---")

	// CASE 1: Refund 3, Nothing Negotiated (Full Return)
	// Return Qty: 3
	// Expected Max: $3,750.00
	// Negotiated: $3,750.00
	t.Run("1. Full Return (3 items)", func(t *testing.T) {
		qty := int64(3)
		negotiated := money.New(375000, "USD")

		res, err := refund.CalculateItemizedRefund(line, qty, negotiated)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		printTwoLevel("1. Full Return", qty, res)
	})

	// CASE 2: Refund 1, Nothing Negotiated (Standard Single Item Return)
	// Return Qty: 1
	// Expected Max: $1,250.00 ($3750 / 3)
	// Negotiated: $1,250.00
	t.Run("2. Single Item Return", func(t *testing.T) {
		qty := int64(1)
		// User asks: "What is the max I can refund for 1 item?"
		// We calculate that implicitly in the function, but in a UI you'd call a helper.
		// For this test, we know it's $1,250.
		negotiated := money.New(125000, "USD")

		res, err := refund.CalculateItemizedRefund(line, qty, negotiated)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		printTwoLevel("2. Single Item", qty, res)
	})

	// CASE 3: Refund 3, Negotiated 30% (Defective but kept)
	// Return Qty: 3
	// Max: $3,750.00
	// Negotiated: $1,125.00 (30%)
	t.Run("3. Qty 3, Negotiated 30%", func(t *testing.T) {
		qty := int64(3)
		negotiated := money.New(112500, "USD")

		res, err := refund.CalculateItemizedRefund(line, qty, negotiated)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		printTwoLevel("3. Negotiated 30%", qty, res)
	})

	// CASE 4: Refund 3, Negotiated 67% (Major damage)
	// Return Qty: 3
	// Max: $3,750.00
	// Negotiated: $2,512.50
	t.Run("4. Qty 3, Negotiated 67%", func(t *testing.T) {
		qty := int64(3)
		negotiated := money.New(251250, "USD")

		res, err := refund.CalculateItemizedRefund(line, qty, negotiated)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		printTwoLevel("4. Negotiated 67%", qty, res)
	})

	// CASE 5: Refund 3, Negotiated $666 (Specific Amount)
	// Return Qty: 3
	// Max: $3,750.00
	// Negotiated: $666.00
	t.Run("5. Qty 3, Negotiated $666", func(t *testing.T) {
		qty := int64(3)
		negotiated := money.New(66600, "USD")

		res, err := refund.CalculateItemizedRefund(line, qty, negotiated)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		printTwoLevel("5. Negotiated $666", qty, res)
	})

	// CASE 6: The Validation Check (Refund 1 item, but try to get $2000 back)
	// Return Qty: 1
	// Max: $1,250.00
	// Negotiated: $2,000.00 (Should Fail)
	t.Run("6. Over-Refund Attempt", func(t *testing.T) {
		qty := int64(1)
		negotiated := money.New(200000, "USD") // $2,000

		_, err := refund.CalculateItemizedRefund(line, qty, negotiated)
		if err == nil {
			t.Errorf("Expected error for over-refund, got nil")
		} else {
			fmt.Printf("6. Over-Refund Check: PASSED (Blocked: %v)\n", err)
		}
	})
}

func printTwoLevel(title string, qty int64, res refund.RefundCalculation) {
	fmt.Printf("%s (Qty: %d)\n", title, qty)
	fmt.Printf("   Max Refundable: %s\n", res.MaxRefundable)
	totalRef := int64(0)
	for _, p := range res.RefundedParts {
		fmt.Printf("   - %s Refund: %s\n", p.Name, p.Amount)
		totalRef += p.Amount.Amount()
	}
	fmt.Printf("   Total Actual Refund: %s\n\n", money.New(totalRef, "USD"))
}
