package billing_test

import (
	"testing"
	"time"

	"github.com/LordAldi/gmoney/pkg/billing"
	"github.com/LordAldi/gmoney/pkg/calendar"
	"github.com/LordAldi/gmoney/pkg/money"
)

// Helper to make dates readable in the table
func date(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestCalculateProratedCharges(t *testing.T) {
	// Setup a standard Mon-Fri work policy
	policy := calendar.NewStandardPolicy()

	tests := []struct {
		name           string
		totalAmount    int64  // The full subscription price
		subStart       string // When the user subscription started
		subEnd         string // When the user subscription ended
		periodStart    string // The billing cycle start
		periodEnd      string // The billing cycle end
		expectedAmount int64  // How much we should charge
	}{
		{
			name:           "Full Cycle Coverage",
			totalAmount:    10000, // $100.00
			subStart:       "2023-01-01",
			subEnd:         "2023-12-31",
			periodStart:    "2023-06-01",
			periodEnd:      "2023-06-30",
			expectedAmount: 10000, // Active for full billing period -> Full Price
		},
		{
			name:        "Mid-Month Join (Halfway)",
			totalAmount: 10000, // $100.00
			// June 2023 has 22 Business Days.
			// Cycle: June 1 (Thu) to June 30 (Fri).
			// User joins June 15 (Thu).
			// Business days in full June: 22
			// Business days June 15-30: 12
			// Ratio: 12 / 22 = 0.5454...
			// Calc: 10000 * 12 / 22 = 5454.54... -> 5455 (Round Half Up/Largest Remainder)
			subStart:       "2023-06-15",
			subEnd:         "2023-12-31",
			periodStart:    "2023-06-01",
			periodEnd:      "2023-06-30",
			expectedAmount: 5455,
		},
		{
			name:        "Leap Year February (Partial)",
			totalAmount: 2100, // $21.00 ($1/day assuming 21 business days)
			// Feb 2024 (Leap) has 21 Business Days.
			// User active Feb 1 to Feb 2 (Thu, Fri) -> 2 Days.
			// Charge: 2 / 21 of $21.00 = $2.00
			subStart:       "2024-02-01",
			subEnd:         "2024-02-02",
			periodStart:    "2024-02-01",
			periodEnd:      "2024-02-29",
			expectedAmount: 200,
		},
		{
			name:        "Weekend Join (Start on Saturday)",
			totalAmount: 10000,
			// User joins Sat June 3rd.
			// Billing starts June 1st.
			// Period June 1 - June 9 (7 Business Days: Thu, Fri, | Sat, Sun | Mon, Tue, Wed, Thu, Fri).
			// User active: Sat June 3 - June 9.
			// User missed Thu, Fri (2 days).
			// Active days: Mon, Tue, Wed, Thu, Fri (5 days).
			// Total Days: 7.
			// Charge: 5/7 of 10000 = 7142.8 -> 7143
			subStart:       "2023-06-03",
			subEnd:         "2023-06-30",
			periodStart:    "2023-06-01",
			periodEnd:      "2023-06-09",
			expectedAmount: 7143,
		},
		{
			name:           "Out of Bounds (Churned user)",
			totalAmount:    10000,
			subStart:       "2023-01-01",
			subEnd:         "2023-05-31", // Cancelled before period starts
			periodStart:    "2023-06-01",
			periodEnd:      "2023-06-30",
			expectedAmount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sub := billing.Subscription{
				TotalAmount: money.New(tc.totalAmount, "USD"),
				Start:       date(tc.subStart),
				End:         date(tc.subEnd),
			}

			charge, err := billing.CalculateProratedCharges(
				sub,
				date(tc.periodStart),
				date(tc.periodEnd),
				policy,
			)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if charge.Amount() != tc.expectedAmount {
				t.Errorf("Expected charge %d, got %d", tc.expectedAmount, charge.Amount())
			}
		})
	}
}
