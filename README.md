# gmoney: Precision Financial Primitives for Go

**gmoney** is a zero-dependency Go library designed for financial systems where "close enough" is not acceptable. It provides primitives for monetary arithmetic, tax handling, proration, tiered pricing, complex refund negotiation, and double-entry accounting without floating-point errors.

> **The Problem:** `100 / 3` in standard math is `33.333...`. In finance, you cannot destroy the remaining `0.001`. You must distribute it. This library handles the "Penny Variance" problem correctly using the Hamilton/Largest Remainder Method, ensuring your general ledger always balances.

## 📦 Features

* **🛡️ Immutable Money:** Safe `int64` wrapper preventing currency mixing (e.g., adding USD to EUR panics or errors).
* **⚖️ Penny-Perfect Allocation:** Split funds by weights (e.g., 33% / 33% / 34%) without losing a single cent. Supports hierarchical (Tree) distribution.
* **📅 O(1) Business Calendar:** Calculate billable days between dates excluding weekends and holidays in constant time (no loops), verified against brute-force logic.
* **📉 Tiered Pricing:** Calculate costs for volume pricing (e.g., "First 10k units @ $0.05, Next 50k @ $0.04") with micro-penny precision.
* **💸 Negotiated Refunds:** Two-level refund engine that handles quantity returns limits and prorates negotiated settlements while preserving tax/base ratios.
* **📖 Double-Entry Ledger:** Transaction engine that enforces , preventing money creation/destruction.
* **💱 Multi-Currency FX:** Settlement engine that calculates Realized Gain/Loss when exchange rates fluctuate between Invoice and Payment dates.

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

// Subscription: $100.00/month. Joined Feb 14-29.
sub := billing.Subscription{
    TotalAmount: money.New(10000, "USD"),
    Start:       time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
    End:         time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
}

// Calculate for February 2024 (Leap Year)
charge, _ := billing.CalculateProratedCharges(sub, startOfFeb, endOfFeb, policy)

```

### 3. Complex Refund Negotiation

Handle the edge case: `I bought 3 items for $3,750. I am returning 1, but we negotiated a custom refund of $666 for the remaining damaged ones.*`

The engine calculates the **Max Refundable Cap** based on quantity, then distributes the negotiated amount across Base Price and Tax to ensure audit compliance.

```go
import "github.com/LordAldi/gmoney/pkg/refund"

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

### 4. Double-Entry Ledger

The Ledger package enforces the **Double-Entry Rule**. It is impossible to create a transaction that destroys or creates money cleanly.

```go
import "github.com/LordAldi/gmoney/pkg/ledger"

// Attempting to book an unbalanced invoice
entries := []ledger.Entry{
    {AccountID: "Receivable", Amount: money.New(1500, "USD")}, // +$15.00
    {AccountID: "Revenue",    Amount: money.New(-1000, "USD")}, // -$10.00
    // Missing -$5.00 Tax!
}

txn, err := ledger.NewTransaction("TXN:1", "Inv#1", entries)
// err: "transaction unbalanced: diff 500 cents"
// The system refuses to record this.

```

### 5. Multi-Currency Settlement (FX)

Calculate **Realized Gain/Loss** when exchange rates fluctuate between the Invoice Date and the Payment Date.

```go
import "github.com/LordAldi/gmoney/pkg/exchange"

// Day 1: Invoiced €100. Booked as $110 USD (Rate 1.10).
bookedAR := money.New(11000, "USD") 

// Day 30: Customer pays €100. Rate is now 1.05.
payment := money.New(10000, "EUR")
currentRate, _ := exchange.NewRate("EUR", "USD", "1.05")

res, _ := exchange.SettlePayment(payment, currentRate, bookedAR)
// res.ConvertedAmount: $105.00
// res.GainLoss:        $5.00
// res.IsGain:          false (Loss)

```

## 🚀 Full Lifecycle Example: Sale, Refund & Ledger

This example demonstrates the complete financial loop: Calculating Tax, Booking Revenue, Calculating a Refund, and Reversing the entries.

```go
func ExampleFullLifecycle() {
    // 1. THE SALE: $2,000 Base + 10% Tax
    basePrice := money.New(200000, "USD")
    taxRate, _ := rate.New("0.10")
    saleResult, _ := tax.CalculateExclusive(basePrice, taxRate) // Total $2,200

    // Book Sale (Dr AR, Cr Sales, Cr Tax)
    ledger.NewTransaction("TXN:101", "Sale", []ledger.Entry{
        {AccountID: "Assets:AR", Amount: saleResult.Total},            // +2200
        {AccountID: "Rev:Sales", Amount: saleResult.Base.Negate()},    // -2000
        {AccountID: "Liab:Tax",  Amount: saleResult.Tax.Negate()},     // -200
    })

    // 2. THE REFUND: Negotiated $660 return
    // Use Refund Engine to split $660 into Base ($600) and Tax ($60)
    orig := []refund.Component{{Name: "Base", Amount: saleResult.Base}, {Name: "Tax", Amount: saleResult.Tax}}
    refundRes, _ := refund.CalculateNegotiatedRefund(orig, money.New(66000, "USD"))

    // Book Refund (Cr AR, Dr Sales, Dr Tax)
    ledger.NewTransaction("TXN:102", "Refund", []ledger.Entry{
        {AccountID: "Assets:AR", Amount: refundRes.Total.Negate()},    // -660 (Reduce Debt)
        {AccountID: "Rev:Sales", Amount: refundRes.Components[0].Amount}, // +600 (Reduce Rev)
        {AccountID: "Liab:Tax",  Amount: refundRes.Components[1].Amount}, // +60  (Reduce Liab)
    })
    
    // Result: Books are balanced. Net Tax Liability matches Net Sales.
}

```

## ⚡ Benchmarks

Core algorithms are optimized for high-frequency trading or billing systems.

```text
BenchmarkCalculateProratedCharges/1_Month-16     2166 ns/op    0 allocs/op
BenchmarkCalculateProratedCharges/1_Year-16      3000 ns/op    0 allocs/op

```

*Note: The Business Day counter uses O(1) math logic with a fallback safety loop, ensuring performance remains constant regardless of the time window size. All keys are integer-based to ensure zero GC pressure.*

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