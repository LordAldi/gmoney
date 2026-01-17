package pricing_test

import (
	"testing"

	"github.com/LordAldi/gmoney/pkg/pricing"
	"github.com/LordAldi/gmoney/pkg/rate"
)

func TestGraduatedPricing_Complex(t *testing.T) {
	// Scenario: A cloud API billing model.
	// 0 - 100 requests: $0.050 (High cost)
	// 100 - 200 requests: $0.045 (Slight discount)
	// 200+ requests: $0.004 (Micro-penny volume pricing)

	r1, _ := rate.New("0.05")
	r2, _ := rate.New("0.045")
	r3, _ := rate.New("0.004")

	tiers := []pricing.Tier{
		{UpTo: 100, Price: r1}, // Capacity 100
		{UpTo: 100, Price: r2}, // Capacity 100 (covering 101-200)
		{UpTo: 0, Price: r3},   // Capacity Infinite
	}

	// Usage: 1250 units
	// Tier 1: 100 * 0.05 = $5.00
	// Tier 2: 100 * 0.045 = $4.50
	// Tier 3: 1050 * 0.004 = $4.20
	// Total Expected: $13.70 (1370 cents)

	usage := int64(1250)
	bill, err := pricing.CalculateGraduatedCost(usage, tiers)
	if err != nil {
		t.Fatalf("Calculation error: %v", err)
	}

	if bill.Amount() != 1370 {
		t.Errorf("Expected 1370 cents ($13.70), got %d cents", bill.Amount())
	}
}

func TestGraduatedPricing_Rounding(t *testing.T) {
	// Scenario: 1 unit at $0.004
	// Total $0.004. Rounds to $0.00 (0 cents).

	// Scenario: 2 units at $0.004
	// Total $0.008. Rounds to $0.01 (1 cent).

	r, _ := rate.New("0.004")
	tiers := []pricing.Tier{{UpTo: 0, Price: r}}

	// Case 1
	bill1, _ := pricing.CalculateGraduatedCost(1, tiers)
	if bill1.Amount() != 0 {
		t.Errorf("1 unit: expected 0 cents, got %d", bill1.Amount())
	}

	// Case 2
	bill2, _ := pricing.CalculateGraduatedCost(2, tiers)
	if bill2.Amount() != 1 {
		t.Errorf("2 units: expected 1 cent (round up from 0.008), got %d", bill2.Amount())
	}
}
