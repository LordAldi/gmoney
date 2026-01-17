package examples_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/LordAldi/gmoney/pkg/billing"
	"github.com/LordAldi/gmoney/pkg/calendar"
	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/pricing"
	"github.com/LordAldi/gmoney/pkg/rate"
	"github.com/LordAldi/gmoney/pkg/tax"
)

// Helper to parse dates quickly
func date(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestGenerateFullInvoice(t *testing.T) {
	// ==========================================
	// 1. SETUP: Define Policies & Rates
	// ==========================================

	// A. Calendar Policy (Standard Mon-Fri)
	policy := calendar.NewStandardPolicy()
	// Add a Bank Holiday (e.g., Presidents' Day if US, or just a test holiday)
	policy.AddHoliday(date("2024-02-19"))

	// B. Usage Rates (Tiered)
	// Tier 1: 0 - 100,000 @ $0.005
	rate1, _ := rate.New("0.005")
	// Tier 2: 100,000+    @ $0.004
	rate2, _ := rate.New("0.004")

	tiers := []pricing.Tier{
		{UpTo: 100_000, Price: rate1},
		{UpTo: 0, Price: rate2}, // 0 = Infinite
	}

	// C. Tax Rate (20% VAT)
	vatRate, _ := rate.New("0.20")

	// ==========================================
	// 2. INPUT: Customer Data
	// ==========================================

	subscriptionPrice := money.New(100000, "USD") // $1,000.00

	sub := billing.Subscription{
		TotalAmount: subscriptionPrice,
		Start:       date("2024-02-14"), // Joined mid-month
		End:         date("2024-02-29"), // Leap year end
	}

	usageCount := int64(150_000) // 150k requests

	billingStart := date("2024-02-01")
	billingEnd := date("2024-02-29")

	// ==========================================
	// 3. EXECUTION: Calculate the Bill
	// ==========================================

	fmt.Println("--- INVOICE CALCULATION LOG ---")

	// STEP A: Calculate Prorated Subscription
	// Uses: pkg/calendar (Business Days) + pkg/allocate (Splitting)
	subCharge, err := billing.CalculateProratedCharges(sub, billingStart, billingEnd, policy)
	if err != nil {
		t.Fatalf("Proration failed: %v", err)
	}
	fmt.Printf("1. Subscription (Prorated): %s\n", subCharge)

	// STEP B: Calculate Usage Costs
	// Uses: pkg/pricing (Tiered Math) + pkg/rate (High Precision)
	// Math Check:
	// 100k * 0.005 = $500.00
	// 50k  * 0.004 = $200.00
	// Total Usage: $700.00
	usageCharge, err := pricing.CalculateGraduatedCost(usageCount, tiers)
	if err != nil {
		t.Fatalf("Usage calc failed: %v", err)
	}
	fmt.Printf("2. Usage Cost (150k reqs):  %s\n", usageCharge)

	// STEP C: Subtotal
	// Uses: pkg/money (Currency Safety)
	subtotal, err := subCharge.Add(usageCharge)
	if err != nil {
		t.Fatalf("Subtotal failed: %v", err)
	}
	fmt.Printf("3. Subtotal:                %s\n", subtotal)

	// STEP D: Tax Calculation
	// Uses: pkg/tax (Exclusive Logic)
	taxRes, err := tax.CalculateExclusive(subtotal, vatRate)
	if err != nil {
		t.Fatalf("Tax calc failed: %v", err)
	}
	fmt.Printf("4. VAT (20%%):               %s\n", taxRes.Tax)
	fmt.Printf("5. TOTAL DUE:               %s\n", taxRes.Total)

	// ==========================================
	// 4. VERIFICATION: The "Senior" Check
	// ==========================================

	// Let's manually verify the Proration Math to be sure.
	// Feb 2024 (Leap) starts Thu Feb 1. Ends Thu Feb 29.
	// Total Days: 29.
	// Weekends:
	//   3-4, 10-11, 17-18, 24-25 (8 weekend days).
	//   Total Weekdays: 29 - 8 = 21.
	// Holiday: Feb 19 (Mon).
	// Total Business Days in Period = 20.

	// User Active: Feb 14 (Wed) to Feb 29 (Thu).
	//   Feb 14-16 (Wed-Fri): 3 days
	//   Feb 17-18 (Weekend): 0 days
	//   Feb 19 (Holiday):    0 days
	//   Feb 20-23 (Tue-Fri): 4 days
	//   Feb 24-25 (Weekend): 0 days
	//   Feb 26-29 (Mon-Thu): 4 days
	// Total Active Days: 3 + 4 + 4 = 11 days.

	// Ratio: 11 / 20 = 0.55
	// Subscription: $1000.00 * 0.55 = $550.00

	expectedSub := int64(55000) // 55000 cents
	if subCharge.Amount() != expectedSub {
		t.Errorf("Proration Math Mismatch! Expected %d, got %d", expectedSub, subCharge.Amount())
	}

	// Verify Usage
	// 100k * 0.005 = 500
	// 50k * 0.004 = 200
	// Total = 700 USD (70000 cents)
	expectedUsage := int64(70000)
	if usageCharge.Amount() != expectedUsage {
		t.Errorf("Usage Math Mismatch! Expected %d, got %d", expectedUsage, usageCharge.Amount())
	}

	// Verify Subtotal
	// 550 + 700 = 1250 USD
	if subtotal.Amount() != 125000 {
		t.Errorf("Subtotal mismatch")
	}

	// Verify Tax
	// 1250 * 0.20 = 250 USD
	if taxRes.Tax.Amount() != 25000 {
		t.Errorf("Tax mismatch! Expected 250.00, got %s", taxRes.Tax)
	}

	// Verify Total
	// 1250 + 250 = 1500 USD
	if taxRes.Total.Amount() != 150000 {
		t.Errorf("Final Total mismatch")
	}
}
