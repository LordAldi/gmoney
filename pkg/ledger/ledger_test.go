package ledger_test

import (
	"fmt"
	"testing"

	"github.com/LordAldi/gmoney/pkg/ledger"
	"github.com/LordAldi/gmoney/pkg/money"
)

func TestBookInvoiceToLedger(t *testing.T) {
	// 1. Setup Accounts
	acctReceivable := "ACCT:1001" // Asset
	revSubscription := "REV:4001" // Revenue
	revUsage := "REV:4002"        // Revenue
	taxPayable := "LIAB:2001"     // Liability

	// 2. The Invoice Data (From your previous example)
	// Total Due: $1,500.00
	// Sub Revenue: $550.00
	// Usage Revenue: $700.00
	// Tax Collected: $250.00

	// 3. Construct Entries
	// Note: In Signed Ledgers:
	// Assets increase with Debit (+)
	// Revenue/Liability increase with Credit (-)

	entries := []ledger.Entry{
		// DEBIT: The customer owes us full amount
		{AccountID: acctReceivable, Amount: money.New(150000, "USD")},

		// CREDIT: Recognize Subscription Revenue
		{AccountID: revSubscription, Amount: money.New(-55000, "USD")},

		// CREDIT: Recognize Usage Revenue
		{AccountID: revUsage, Amount: money.New(-70000, "USD")},

		// CREDIT: We owe this to the Gov, it's not ours
		{AccountID: taxPayable, Amount: money.New(-25000, "USD")},
	}

	// 4. Attempt to Book
	txn, err := ledger.NewTransaction("TXN:001", "Invoice #Feb-2024", entries)
	if err != nil {
		t.Fatalf("Ledger rejected the transaction: %v", err)
	}

	fmt.Printf("Transaction %s posted successfully.\n", txn.ID)
	fmt.Println("--- GL POSTING ---")
	for _, e := range txn.Entries {
		fmt.Printf("Account %s: %s\n", e.AccountID, e.Amount)
	}
}

func TestBookUnbalancedLedger(t *testing.T) {
	// Scenario: A Junior Dev tries to book revenue but forgets the Tax Liability.
	// Customer pays $1,500, but we only record $1,250 revenue.
	// $250 vanishes into thin air.

	entries := []ledger.Entry{
		{AccountID: "ACCT:1001", Amount: money.New(150000, "USD")}, // +1500
		{AccountID: "REV:4001", Amount: money.New(-125000, "USD")}, // -1250
		// Missing -250 entry
	}

	_, err := ledger.NewTransaction("TXN:002", "Bad Invoice", entries)
	if err == nil {
		t.Fatal("Ledger should have rejected unbalanced transaction!")
	}

	fmt.Printf("\nSafety Check Passed: %v\n", err)
}

func TestNewTransaction_BalancesCorrectly(t *testing.T) {
	// Scenario: Standard Invoice Booking
	// Customer owes us $1500 (Debit +1500)
	// We recognize Revenue $1250 (Credit -1250)
	// We owe Tax $250 (Credit -250)
	// Sum: 1500 + (-1250) + (-250) = 0.

	entries := []ledger.Entry{
		{AccountID: "ACCT:RECEIVABLE", Amount: money.New(150000, "USD")},
		{AccountID: "REV:SUBSCRIPTION", Amount: money.New(-125000, "USD")},
		{AccountID: "LIAB:TAX_PAYABLE", Amount: money.New(-25000, "USD")},
	}

	txn, err := ledger.NewTransaction("TXN:001", "Invoice #1", entries)
	if err != nil {
		t.Fatalf("Expected valid transaction, got error: %v", err)
	}

	if len(txn.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(txn.Entries))
	}
}

func TestNewTransaction_RejectsUnbalanced(t *testing.T) {
	// Scenario: Missing the Tax entry.
	// Debit +1500, Credit -1250. Diff +250.
	entries := []ledger.Entry{
		{AccountID: "ACCT:RECEIVABLE", Amount: money.New(150000, "USD")},
		{AccountID: "REV:SUBSCRIPTION", Amount: money.New(-125000, "USD")},
	}

	_, err := ledger.NewTransaction("TXN:002", "Bad Invoice", entries)

	// We EXPECT an error here.
	if err == nil {
		t.Fatal("Ledger failed to catch unbalanced transaction!")
	}

	expectedErr := "transaction does not balance! Diff: 25000 cents"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestNewTransaction_RejectsMixedCurrency(t *testing.T) {
	entries := []ledger.Entry{
		{AccountID: "ACCT:US_BANK", Amount: money.New(100, "USD")},
		{AccountID: "ACCT:EU_BANK", Amount: money.New(-100, "EUR")},
	}

	_, err := ledger.NewTransaction("TXN:003", "Forex Hack", entries)
	if err == nil {
		t.Fatal("Ledger allowed mixed currencies!")
	}
}
