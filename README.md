# gmoney: Precision Financial Primitives for Go

**gmoney** is a zero-dependency Go library designed for financial systems where "close enough" is not acceptable. It provides primitives for monetary arithmetic, tax handling, proration, and high-precision allocation without floating-point errors.

> **The Problem:** `100 / 3` in standard math is `33.333...`. In finance, you cannot destroy the remaining `0.001`. You must distribute it. This library handles the "Penny Variance" problem correctly using the Hamilton/Largest Remainder Method.

## 📦 Features

* **🛡️ Immutable Money:** Safe `int64` wrapper preventing currency mixing (e.g., adding USD to EUR panics or errors).
* **⚖️ Penny-Perfect Allocation:** Split funds by weights (e.g., 33% / 33% / 34%) without losing a single cent. Supports hierarchical (Tree) distribution.
* **📅 O(1) Business Calendar:** Calculate billable days between dates excluding weekends and holidays in constant time (no loops), verified against brute-force logic.
* **📉 Tiered Pricing:** Calculate costs for volume pricing (e.g., "First 10k units @ $0.05, Next 50k @ $0.04") with micro-penny precision.
* **🏛️ Tax Engine:** Handle Inclusive (VAT) and Exclusive (Sales Tax) calculations without rounding drift.
* **🧪 Property-Based Tested:** Logic verified with thousands of random inputs via `gopter` to prove invariants hold (e.g., Conservation of Money).

## 🚀 Installation

```bash
go get github.com/LordAldi/gmoney

```

## 💡 Quick Start

### 1. The "Penny Split" (Allocation)

Distribute $0.05 among 3 people equally. Standard division gives 1.66 cents. **gmoney** distributes the remainder fairly.

```go
package main

import (
	"fmt"
	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/allocate"
)

func main() {
	total := money.New(5, "USD") // 5 cents
	weights := []int{1, 1, 1}    // Equal split

	parts, _ := allocate.Split(total, weights)

	// Result: [2, 2, 1] (Total sums to 5. No money destroyed.)
	fmt.Println(parts) 
}

```

### 2. Proration (Billing)

Charge a customer for exactly the number of *business days* they used a service in a leap year.

```go
import (
	"github.com/LordAldi/gmoney/pkg/billing"
	"github.com/LordAldi/gmoney/pkg/calendar"
)

// Setup policy (Mon-Fri, with custom holidays)
policy := calendar.NewStandardPolicy()
policy.AddHoliday(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

// Subscription: $100.00/month
sub := billing.Subscription{
    TotalAmount: money.New(10000, "USD"),
    Start:       time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
    End:         time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
}

// Calculate for February 2024 (Leap Year)
charge, _ := billing.CalculateProratedCharges(sub, startOfFeb, endOfFeb, policy)

```

### 3. Inclusive Tax (VAT)

Extract the base price from a gross total of £10.00 with 20% VAT.

```go
import (
	"github.com/LordAldi/gmoney/pkg/tax"
	"github.com/LordAldi/gmoney/pkg/rate"
)

gross := money.New(1000, "GBP") // £10.00
vatRate, _ := rate.New("0.20")

// Math: 1000 / 1.20 = 833.333...
res, _ := tax.CalculateInclusive(gross, vatRate)

fmt.Println(res.Base) // 833 (£8.33)
fmt.Println(res.Tax)  // 167 (£1.67)
// 833 + 167 = 1000. Correct.

```


### 4. Complex Refund Negotiation

Handle the edge case: `I bought 3 items for $3,750. I am returning 1, but we negotiated a custom refund of $666 for the remaining damaged ones.*`

The engine calculates the **Max Refundable Cap** based on quantity, then distributes the negotiated amount across Base Price and Tax to ensure audit compliance.

```go
import "github.com/your-username/gmoney/pkg/refund"

// Original Line Item: 3 Units, Total $3,750 ($3000 Base + $750 Tax)
line := refund.LineItem{
    Quantity: 3,
    Components: []refund.Component{
        {Name: "Base", Amount: money.New(300000, "USD")},
        {Name: "Tax",  Amount: money.New(75000, "USD")},
    },
}

// Scenario: Return 3 items, but negotiated $666.00 refund
res, err := refund.CalculateItemizedRefund(line, 3, money.New(66600, "USD"))

// Result preserves the 80/20 Tax Ratio of the original order:
// Max Refundable: $3,750.00
// Base Refund:    $532.80
// Tax Refund:     $133.20
// Total:          $666.00

```

### 5 🚀 Real-World Scenario: The "Cloud Invoice"

Combine all packages to generate a complex SaaS invoice:

* **Scenario:** Customer joined Feb 14th (Leap Year).
* **Usage:** 150k API requests (Tiered pricing).
* **Logic:** Prorate subscription by *business days* (excluding Feb 19th Holiday), calculate tiered usage costs, then apply 20% VAT.

```go
func ExampleGenerateInvoice() {
	// 1. Setup Policies
	policy := calendar.NewStandardPolicy()
	policy.AddHoliday(time.Date(2024, 2, 19, 0, 0, 0, 0, time.UTC)) // Bank Holiday

	// 2. Define Pricing Tiers
	// First 100k @ $0.005, Rest @ $0.004
	rate1, _ := rate.New("0.005")
	rate2, _ := rate.New("0.004")
	tiers := []pricing.Tier{
		{UpTo: 100_000, Price: rate1},
		{UpTo: 0,       Price: rate2},
	}

	// 3. Calculate Subscription (Prorated)
	// $1,000/month. Joined Feb 14-29 (11 active business days / 20 total)
	sub := billing.Subscription{
		TotalAmount: money.New(100000, "USD"),
		Start:       time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
		End:         time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
	}
	subCharge, _ := billing.CalculateProratedCharges(sub, startFeb, endFeb, policy)
	// Result: $550.00

	// 4. Calculate Usage
	// 100k * 0.005 ($500) + 50k * 0.004 ($200)
	usageCharge, _ := pricing.CalculateGraduatedCost(150_000, tiers)
	// Result: $700.00

	// 5. Total & Tax
	subtotal, _ := subCharge.Add(usageCharge) // $1,250.00
	vatRate, _ := rate.New("0.20")
	finalBill, _ := tax.CalculateExclusive(subtotal, vatRate)

	fmt.Printf("Subtotal: %s\n", finalBill.Base)  // 125000 ($1,250.00)
	fmt.Printf("Tax:      %s\n", finalBill.Tax)   // 25000  ($250.00)
	fmt.Printf("Total:    %s\n", finalBill.Total) // 150000 ($1,500.00)
}

```

## ⚡ Benchmarks

Core algorithms are optimized for high-frequency trading or billing systems.

```text
BenchmarkCalculateProratedCharges/1_Month-16     2166 ns/op    0 allocs/op
BenchmarkCalculateProratedCharges/1_Year-16      3000 ns/op    0 allocs/op

```

*Note: The Business Day counter uses O(1) math logic with a fallback safety loop, ensuring performance remains constant regardless of the time window size.*

## 🧪 Verification

This library uses **Property-Based Testing** (`gopter`) to ensure mathematical laws hold true under stress.

```bash
# Run the stress tests
go test ./pkg/allocate -v -run TestSplit_Properties

```

*Checks performed:*

* **Conservation:** Sum of parts must exactly equal the total.
* **Non-Negativity:** No part can be negative if total is positive.
* **Monotonicity:** Larger weights must never receive less money than smaller weights.

## 🤝 Contributing

PRs are welcome. Please ensure that:

1. New logic includes Property Tests.
2. Allocations are kept to zero where possible.

## 📄 License

MIT

