// pkg/exchange/settlement.go
package exchange

import "github.com/LordAldi/gmoney/pkg/money"

type SettlementResult struct {
	ConvertedAmount money.Money // What we actually got in Base Currency
	GainLoss        money.Money // The difference due to rate fluctuation
	IsGain          bool        // True if Gain, False if Loss
}

// SettlePayment calculates the FX impact of receiving foreign currency.
// foreignAmount: The amount received (e.g., €100)
// currentRate:   The rate today (e.g., 1.05)
// originalValue: The value of this money when we booked the invoice (e.g., $110)
func SettlePayment(foreignAmount money.Money, currentRate Rate, originalValue money.Money) (SettlementResult, error) {
	// 1. Convert Foreign to Base using TODAY's rate
	actualValue, err := Convert(foreignAmount, currentRate)
	if err != nil {
		return SettlementResult{}, err
	}

	// 2. Calculate Difference (Actual - Original)
	diff := actualValue.Amount() - originalValue.Amount()

	gainLoss := money.New(diff, originalValue.Currency())
	isGain := diff >= 0

	// If loss, we usually want the magnitude (positive number) for the ledger entry logic
	if diff < 0 {
		gainLoss = money.New(-diff, originalValue.Currency())
	}

	return SettlementResult{
		ConvertedAmount: actualValue,
		GainLoss:        gainLoss,
		IsGain:          isGain,
	}, nil
}
