package refund_test

import (
	"testing"

	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/refund"
)

func TestNegotiatedRefund_PartialWorks(t *testing.T) {
	// Original Invoice: $1250 Total
	// - Item Cost: $1000
	// - Tax:       $250
	original := []refund.Component{
		{Name: "Base", Amount: money.New(100000, "USD")}, // 1000.00
		{Name: "Tax", Amount: money.New(25000, "USD")},   // 250.00
	}

	// Negotiated Refund: $666.00
	refundRequest := money.New(66600, "USD")

	res, err := refund.CalculateNegotiatedRefund(original, refundRequest)
	if err != nil {
		t.Fatalf("Calculation failed: %v", err)
	}

	// --- VERIFICATION ---

	// 1. Total Check
	if res.Total.Amount() != 66600 {
		t.Errorf("Total mismatch")
	}

	// 2. Ratio Check
	// We expect the refund to be split 80% Base / 20% Tax (same as 1000/1250)
	// Base Refund: 666 * 0.80 = 532.80 -> Rounds to 533.00?
	// Tax Refund:  666 * 0.20 = 133.20 -> Rounds to 133.00?
	// Note: 532.80 rounds to 533. 133.20 rounds to 133. Sum = 666. Correct.

	baseRefund := res.Components[0].Amount.Amount()
	taxRefund := res.Components[1].Amount.Amount()

	if baseRefund != 53280 { // $532.80
		t.Errorf("Expected Base Refund $532.80, got %d", baseRefund)
	}
	if taxRefund != 13320 { // $133.20
		t.Errorf("Expected Tax Refund $133.20, got %d", taxRefund)
	}

	t.Logf("Original:  $1000 Base, $250 Tax")
	t.Logf("Refunded:  $%d Base, $%d Tax (Total $%d)", baseRefund/100, taxRefund/100, (baseRefund+taxRefund)/100)
}

func TestCalculateNegotiatedRefund_Scenarios(t *testing.T) {
	// Helper to create Money quickly (cents)
	m := func(amount int64) money.Money {
		return money.New(amount, "USD")
	}

	tests := []struct {
		name         string
		originalBase int64 // The original Base Price (e.g., $1000.00)
		originalTax  int64 // The original Tax (e.g., $250.00)
		refundAmount int64 // The TOTAL amount we agreed to refund
		expectedBase int64 // How much of the refund should be attributed to Base
		expectedTax  int64 // How much of the refund should be attributed to Tax
	}{
		{
			// Context: Bought 3 items for $300 total ($200 Base + $100 Tax).
			// Refund: Full return (Refunding all $300).
			// Math: Returns exactly the original amounts.
			name:         "Full Refund (3 of 3 Qty)",
			originalBase: 20000,
			originalTax:  10000,
			refundAmount: 30000,
			expectedBase: 20000,
			expectedTax:  10000,
		},
		{
			// Context: Bought 3 items for $300. Returning 1 item ($100).
			// Ratio: Base is 66.6% ($200/$300). Tax is 33.3% ($100/$300).
			// Calculation:
			// Refund $100.
			// Base share: 100 * (2/3) = 66.66... -> Rounds to 67 cents ($0.67) ??
			// Wait, let's use larger numbers.
			// Base $200.00, Tax $100.00. Total $300.00.
			// Refund $100.00.
			// Base part: 100.00 * (200/300) = 66.666... -> $66.67
			// Tax part:  100.00 * (100/300) = 33.333... -> $33.33
			// Check: 66.67 + 33.33 = 100.00.
			name:         "Partial Qty Refund (1 of 3 Qty)",
			originalBase: 20000,
			originalTax:  10000,
			refundAmount: 10000,
			expectedBase: 6667, // $66.67
			expectedTax:  3333, // $33.33
		},
		{
			// Context: Item works but is scratched. Negotiated 30% refund.
			// Original: $1000 Base, $250 Tax. Total $1250.
			// Refund: 30% of $1250 = $375.
			// Split Ratio: 80% Base (1000/1250), 20% Tax (250/1250).
			// Base Refund: 375 * 0.80 = 300.
			// Tax Refund:  375 * 0.20 = 75.
			name:         "Negotiated 30% Refund",
			originalBase: 100000,
			originalTax:  25000,
			refundAmount: 37500, // $375.00
			expectedBase: 30000, // $300.00
			expectedTax:  7500,  // $75.00
		},
		{
			// Context: Item is mostly broken. Negotiated 67% refund.
			// Original: $1000 Base, $250 Tax. Total $1250.
			// Refund: 67% of $1250 = $837.50. Let's say $837.50 exactly.
			// Base Refund: 837.50 * 0.80 = 670.00.
			// Tax Refund:  837.50 * 0.20 = 167.50.
			name:         "Negotiated 67% Refund",
			originalBase: 100000,
			originalTax:  25000,
			refundAmount: 83750,
			expectedBase: 67000, // $670.00
			expectedTax:  16750, // $167.50
		},
		{
			// Context: The "Specific Number" Case.
			// Original: $1250 Total ($1000 Base, $250 Tax).
			// Refund: $666.00.
			// Base Refund: 666 * 0.8 = 532.8 -> Rounds to 533 (Half Up? Or Largest Remainder?)
			// Tax Refund:  666 * 0.2 = 133.2 -> Rounds to 133.
			// Sum: 533 + 133 = 666.
			// Note: 532.8 usually rounds to 533. 133.2 rounds to 133. Perfect.
			name:         "The $666 Negotiated Case",
			originalBase: 100000,
			originalTax:  25000,
			refundAmount: 66600,
			expectedBase: 53280, // $532.80
			expectedTax:  13320, // $133.20
		},
		{
			// Context: Extreme Edge Case (1 Penny Refund).
			// Who gets the penny? The largest weight (Base).
			name:         "The 1 Penny Refund",
			originalBase: 100000, // $1000
			originalTax:  25000,  // $250
			refundAmount: 1,      // $0.01
			expectedBase: 1,      // Base gets it (larger weight)
			expectedTax:  0,      // Tax gets nothing
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Construct the input components
			originalParts := []refund.Component{
				{Name: "Base", Amount: m(tc.originalBase)},
				{Name: "Tax", Amount: m(tc.originalTax)},
			}

			// Run the calculation
			res, err := refund.CalculateNegotiatedRefund(originalParts, m(tc.refundAmount))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify Base Share
			if res.Components[0].Amount.Amount() != tc.expectedBase {
				t.Errorf("Base Mismatch! Expected %d, got %d", tc.expectedBase, res.Components[0].Amount.Amount())
			}

			// Verify Tax Share
			if res.Components[1].Amount.Amount() != tc.expectedTax {
				t.Errorf("Tax Mismatch! Expected %d, got %d", tc.expectedTax, res.Components[1].Amount.Amount())
			}

			// Verify Conservation of Money (Base + Tax MUST equal Refund Total)
			totalRefunded := res.Components[0].Amount.Amount() + res.Components[1].Amount.Amount()
			if totalRefunded != tc.refundAmount {
				t.Errorf("Money Lost/Gained! Refunded %d, expected %d", totalRefunded, tc.refundAmount)
			}
		})
	}
}
