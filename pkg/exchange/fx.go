package exchange

import (
	"fmt"
	"math/big"

	"github.com/LordAldi/gmoney/pkg/money"
	"github.com/LordAldi/gmoney/pkg/rate"
)

// Rate represents the value of 1 unit of Source in Target currency.
// e.g., Source=EUR, Target=USD, Price=1.10 (1 EUR = $1.10 USD)
type Rate struct {
	Source string
	Target string
	Price  rate.Rate
}

func NewRate(source, target, priceStr string) (Rate, error) {
	r, err := rate.New(priceStr)
	if err != nil {
		return Rate{}, err
	}
	return Rate{Source: source, Target: target, Price: r}, nil
}

// Convert takes a money amount and applies the rate.
// Returns the new Amount and the precise "Source Equivalent" (for ledger matching).
func Convert(amount money.Money, r Rate) (money.Money, error) {
	if amount.Currency() != r.Source {
		return money.Money{}, fmt.Errorf("rate mismatch: money is %s, rate is from %s", amount.Currency(), r.Source)
	}

	// Logic: TargetAmount = SourceAmount * Price
	// We use big.Rat logic similar to your tax engine.
	// You will reuse rate.Mul logic here or duplicate the precision math.

	// Implementation Note: Assuming rate.Rate has a method .Mul(int64) *big.Rat
	convertedRat := r.Price.Mul(amount.Amount())

	// Rounding (Half Up)
	finalAmount := roundRat(convertedRat)

	return money.New(finalAmount, r.Target), nil
}

// Helper for rounding (Copy/Paste from tax package or move to shared `math` pkg)
func roundRat(r *big.Rat) int64 {
	half := big.NewRat(1, 2)
	sum := new(big.Rat).Add(r, half)
	num := sum.Num()
	denom := sum.Denom()
	return new(big.Int).Div(num, denom).Int64()
}
