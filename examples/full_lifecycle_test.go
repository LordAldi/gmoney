package examples_test

import (
	"fmt"
	"testing"

	"github.com/LordAldi/gmoney/pkg/ledger"
	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/rate"
	"github.com/LordAldi/gmoney/pkg/refund"
	"github.com/LordAldi/gmoney/pkg/tax"
)

func TestFullLifecycle_SaleAndRefund(t *testing.T) {
	fmt.Println("--- FULL FINANCIAL LIFECYCLE LOG ---")

	// ==========================================
	// PART 1: THE SALE (Calculation + Booking)
	// ==========================================

	// Scenario: Selling a high-end laptop for $2,000.00 + 10% Tax.
	basePrice := money.New(200000, "USD")
	taxRate, _ := rate.New("0.10")

	// 1. Calculate Tax
	saleResult, _ := tax.CalculateExclusive(basePrice, taxRate)

	fmt.Printf("1. SALE CALCULATION:\n")
	fmt.Printf("   Base: %s | Tax: %s | Total: %s\n", saleResult.Base, saleResult.Tax, saleResult.Total)

	// 2. Book the Sale to Ledger
	// Debit: Accounts Receivable (Asset) +$2,200
	// Credit: Sales Revenue (Revenue)    -$2,000
	// Credit: Tax Payable (Liability)    -$200
	saleEntries := []ledger.Entry{
		{AccountID: "Assets:Receivable", Amount: saleResult.Total}, // +2200.00
		{AccountID: "Rev:Sales", Amount: saleResult.Base.Negate()}, // -2000.00
		{AccountID: "Liab:Tax", Amount: saleResult.Tax.Negate()},   // -200.00
	}

	saleTxn, err := ledger.NewTransaction("TXN:101", "Inv #101", saleEntries)
	if err != nil {
		t.Fatalf("Failed to book sale: %v", err)
	}
	fmt.Printf("   [Ledger] Transaction %s posted. (Reference: %s)\n\n", saleTxn.ID, saleTxn.Reference)

	// ==========================================
	// PART 2: THE NEGOTIATED REFUND
	// ==========================================

	// Scenario: Customer complains about a scratch. We agree to refund $660.00.
	// We must split this $660 into Base and Tax so our books stay accurate.

	// 1. Prepare Inputs for Refund Engine
	originalLines := []refund.Component{
		{Name: "Base", Amount: saleResult.Base}, // $2000
		{Name: "Tax", Amount: saleResult.Tax},   // $200
	}
	negotiatedAmount := money.New(66000, "USD") // $660.00

	// 2. Calculate Refund Split
	refundResult, err := refund.CalculateNegotiatedRefund(originalLines, negotiatedAmount)
	if err != nil {
		t.Fatalf("Refund calc failed: %v", err)
	}

	// Extract the split parts
	baseRefund := refundResult.Components[0].Amount // Should be $600
	taxRefund := refundResult.Components[1].Amount  // Should be $60

	fmt.Printf("2. REFUND CALCULATION (Negotiated $660):\n")
	fmt.Printf("   Base Refund: %s\n", baseRefund)
	fmt.Printf("   Tax Refund:  %s\n", taxRefund)

	// ==========================================
	// PART 3: BOOKING THE REFUND
	// ==========================================

	// Logic: We are giving money back.
	// Credit: Assets:Receivable (Reducing what they owe us) -$660
	// Debit:  Rev:Sales (Contra-Revenue, reducing sales)    +$600
	// Debit:  Liab:Tax (Reducing what we owe the gov)       +$60

	refundEntries := []ledger.Entry{
		// We reduce the asset (Credit)
		{AccountID: "Assets:Receivable", Amount: negotiatedAmount.Negate()}, // -660.00

		// We reduce the revenue (Debit)
		{AccountID: "Rev:Sales", Amount: baseRefund}, // +600.00

		// We reduce the tax liability (Debit)
		{AccountID: "Liab:Tax", Amount: taxRefund}, // +60.00
	}

	refundTxn, err := ledger.NewTransaction("TXN:102", "Refund #101-A", refundEntries)
	if err != nil {
		t.Fatalf("Failed to book refund: %v", err)
	}
	fmt.Printf("   [Ledger] Transaction %s posted. (Reference: %s)\n", refundTxn.ID, refundTxn.Reference)

	// ==========================================
	// PART 4: VERIFICATION (The "T-Account" Check)
	// ==========================================

	// Let's check the final balance of "Liab:Tax".
	// Started with -200 (Credit). Added +60 (Debit). Net should be -140.
	// Meaning: We owe the government $140 on the net $1,400 sale.
	// Math: $1,400 * 10% = $140. Perfect.

	netTaxLiability := saleEntries[2].Amount.Amount() + refundEntries[2].Amount.Amount()
	expectedLiability := int64(-14000) // -$140.00

	if netTaxLiability != expectedLiability {
		t.Errorf("Accounting Drift Detected! Expected Tax Liability %d, got %d", expectedLiability, netTaxLiability)
	} else {
		fmt.Printf("\n   [Audit] Tax Liability Account Balance: %s (Correct)\n", money.New(netTaxLiability, "USD"))
	}
}
