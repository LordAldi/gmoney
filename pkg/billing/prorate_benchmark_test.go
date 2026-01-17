package billing_test

import (
	"testing"

	"github.com/LordAldi/gmoney/pkg/billing"
	"github.com/LordAldi/gmoney/pkg/calendar"
	"github.com/LordAldi/gmoney/pkg/money"
)

// Prevent compiler optimization by assigning result to a global variable
var benchmarkResult money.Money

func BenchmarkCalculateProratedCharges(b *testing.B) {
	// 1. Setup Shared State (Allocated ONCE outside the loop)
	policy := calendar.NewStandardPolicy()

	// Add a few holidays to force map lookups
	policy.AddHoliday(date("2023-12-25"))
	policy.AddHoliday(date("2024-01-01"))

	subAmount := money.New(10000, "USD")

	b.Run("1_Month_Window", func(b *testing.B) {
		sub := billing.Subscription{
			TotalAmount: subAmount,
			Start:       date("2023-06-15"), // Mid-month join
			End:         date("2023-12-31"),
		}
		pStart := date("2023-06-01")
		pEnd := date("2023-06-30")

		b.ReportAllocs() // crucial for tracking Garbage Collection pressure
		b.ResetTimer()   // Don't include setup time in the metrics

		for i := 0; i < b.N; i++ {
			// We capture the output to 'benchmarkResult' to ensure the compiler
			// doesn't optimize the function call away entirely.
			res, _ := billing.CalculateProratedCharges(sub, pStart, pEnd, policy)
			benchmarkResult = res
		}
	})

	b.Run("1_Year_Window", func(b *testing.B) {
		// This stresses the loop in CountBusinessDays (approx 260 iterations)
		sub := billing.Subscription{
			TotalAmount: subAmount,
			Start:       date("2023-01-01"),
			End:         date("2024-01-01"),
		}
		pStart := date("2023-01-01")
		pEnd := date("2023-12-31")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, _ := billing.CalculateProratedCharges(sub, pStart, pEnd, policy)
			benchmarkResult = res
		}
	})
}
