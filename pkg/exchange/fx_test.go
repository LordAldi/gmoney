package exchange_test

import (
	"fmt"
	"testing"

	"github.com/LordAldi/gmoney/pkg/exchange"
	"github.com/LordAldi/gmoney/pkg/ledger"
	"github.com/LordAldi/gmoney/pkg/money"
)

func TestFXSettlement_Loss(t *testing.T) {
	fmt.Println("--- FX SETTLEMENT LOG ---")

	// ==========================================
	// 1. INVOICING (Day 1)
	// ==========================================
	// We billed €100. Rate was 1.10 USD/EUR.
	// We booked Accounts Receivable as $110.00 USD.
	bookedReceivable := money.New(11000, "USD")

	fmt.Printf("1. Invoice Booked: %s (Represents €100 @ 1.10)\n", bookedReceivable)

	// ==========================================
	// 2. PAYMENT (Day 30)
	// ==========================================
	// Customer pays the full €100.
	paymentEur := money.New(10000, "EUR")

	// BUT, the Dollar got stronger. Rate is now 1.05 USD/EUR.
	rateDay30, _ := exchange.NewRate("EUR", "USD", "1.05")

	// Calculate the Settlement
	res, _ := exchange.SettlePayment(paymentEur, rateDay30, bookedReceivable)

	fmt.Printf("2. Payment Received: %s\n", paymentEur)
	fmt.Printf("   Current Value:    %s (Rate 1.05)\n", res.ConvertedAmount)

	if !res.IsGain {
		fmt.Printf("   FX Result:        LOSS of %s\n", res.GainLoss)
	}

	// ==========================================
	// 3. BOOKING THE LEDGER
	// ==========================================
	// We need to zero out the $110 Receivable, even though we only got $105 Cash.
	// The $5 difference is the Expense (Loss).

	entries := []ledger.Entry{
		// 1. CASH (Debit): We got $105 real dollars
		{AccountID: "Asset:Cash", Amount: res.ConvertedAmount}, // +$105.00

		// 2. RECEIVABLE (Credit): We clear the full $110 debt
		// (We must credit the ORIGINAL amount to zero out the account)
		{AccountID: "Asset:AR", Amount: bookedReceivable.Negate()}, // -$110.00

		// 3. FX LOSS (Debit): The balancing plug
		// We lost money, so it's an Expense (Debit)
		{AccountID: "Exp:FX_Loss", Amount: res.GainLoss}, // +$5.00
	}

	// Verify it balances: 105 + (-110) + 5 = 0.
	txn, err := ledger.NewTransaction("TXN:FX-01", "Settlement €100", entries)
	if err != nil {
		t.Fatalf("Ledger unbalanced! %v", err)
	}

	fmt.Printf("3. Ledger Transaction Posted: %s\n", txn.ID)
	for _, e := range txn.Entries {
		fmt.Printf("   %s : %s\n", e.AccountID, e.Amount)
	}
}
