package tax

import (
	"math/big"

	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/rate"
)

// TaxResult holds the split between the base cost and the tax amount.
type TaxResult struct {
	Base  money.Money
	Tax   money.Money
	Total money.Money
}

// CalculateInclusive "Backs out" tax from a gross total.
// Formula: Base = Total / (1 + Rate)
func CalculateInclusive(gross money.Money, taxRate rate.Rate) (TaxResult, error) {
	// 1. Convert Money to Rat (cents)
	grossRat := new(big.Rat).SetInt64(gross.Amount())

	// 2. Prepare (1 + Rate)
	// We need to access the raw big.Rat from your Rate struct.
	// Assuming rate.Rate has a method to return *big.Rat or is exported.
	// For this snippet, let's assume we extended Rate with a raw getter or helper.
	one := big.NewRat(1, 1)
	rateRat := taxRate.Raw()
	denominator := new(big.Rat).Add(one, rateRat)

	// 3. Perform Division: Base = Gross / (1 + Rate)
	baseRat := new(big.Rat).Quo(grossRat, denominator)

	// 4. Rounding (Banker's or Half Up) to get Base Cents
	baseCents := roundRat(baseRat)

	// 5. Calculate Tax Difference (Tax = Gross - Base)
	// This is safer than multiplying, as it guarantees Total == Base + Tax
	taxCents := gross.Amount() - baseCents

	return TaxResult{
		Base:  money.New(baseCents, gross.Currency()),
		Tax:   money.New(taxCents, gross.Currency()),
		Total: gross,
	}, nil
}

// CalculateExclusive adds tax on top of a base price.
// Formula: Tax = Base * Rate
func CalculateExclusive(base money.Money, taxRate rate.Rate) (TaxResult, error) {
	baseRat := new(big.Rat).SetInt64(base.Amount())
	rateRat := taxRate.Raw()

	// Tax = Base * Rate
	taxRat := new(big.Rat).Mul(baseRat, rateRat)
	taxCents := roundRat(taxRat)

	return TaxResult{
		Base:  base,
		Tax:   money.New(taxCents, base.Currency()),
		Total: money.New(base.Amount()+taxCents, base.Currency()),
	}, nil
}

// roundRat rounds a big.Rat to the nearest int64 (Half Up).
func roundRat(r *big.Rat) int64 {
	// Add 0.5
	half := big.NewRat(1, 2)
	sum := new(big.Rat).Add(r, half)

	// Floor by integer division of Num/Denom
	num := sum.Num()
	denom := sum.Denom()
	return new(big.Int).Div(num, denom).Int64()
}
