// pkg/tax/tax_test.go
package tax_test

import (
	"testing"

	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/rate"
	"github.com/LordAldi/gmoney/pkg/tax"
)

func TestCalculateInclusive_VAT(t *testing.T) {
	// Scenario: UK VAT is 20%. Item costs £10.00 (1000 pence).
	// Math: 1000 / 1.2 = 833.333...
	// Expected Base: 833 (£8.33)
	// Expected Tax: 167 (£1.67)
	// Check: 833 + 167 = 1000.

	gross := money.New(1000, "GBP")
	vatRate, _ := rate.New("0.20")

	res, err := tax.CalculateInclusive(gross, vatRate)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if res.Base.Amount() != 833 {
		t.Errorf("Expected Base 833, got %d", res.Base.Amount())
	}
	if res.Tax.Amount() != 167 {
		t.Errorf("Expected Tax 167, got %d", res.Tax.Amount())
	}
	if res.Base.Amount()+res.Tax.Amount() != res.Total.Amount() {
		t.Errorf("Math failure: Base + Tax != Total")
	}
}

func TestCalculateExclusive_NYC(t *testing.T) {
	// Scenario: NYC Sales Tax is 8.875% ($0.08875).
	// Item: $100.00
	// Tax: 100 * 0.08875 = 8.875 -> Rounds up to $8.88

	base := money.New(10000, "USD")
	nycRate, _ := rate.New("0.08875")

	res, _ := tax.CalculateExclusive(base, nycRate)

	if res.Tax.Amount() != 888 {
		t.Errorf("Expected Tax 888, got %d", res.Tax.Amount())
	}
}
