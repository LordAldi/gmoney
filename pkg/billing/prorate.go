package billing

import (
	"time"

	"github.com/LordAldi/gmoney/pkg/allocate"
	"github.com/LordAldi/gmoney/pkg/calendar"
	"github.com/LordAldi/gmoney/pkg/money"
)

type Subscription struct {
	TotalAmount money.Money
	Start       time.Time
	End         time.Time
}

// CalculateProratedCharges splits a subscription fee accurately across two periods:
// 1. The "Active" period (Customer was using the service)
// 2. The "Unused" period (Customer hadn't joined yet or cancelled)
func CalculateProratedCharges(sub Subscription, periodStart, periodEnd time.Time, pol *calendar.Policy) (money.Money, error) {
	// 1. Calculate the total theoretical business days in the billing cycle
	totalDays, err := pol.CountBusinessDays(periodStart, periodEnd)
	if err != nil {
		return money.Money{}, err
	}

	if totalDays == 0 {
		return money.New(0, sub.TotalAmount.Currency()), nil
	}

	// 2. Calculate the customer's active days within that window
	// (Intersection of User Duration AND Billing Period)
	activeStart := maxTime(sub.Start, periodStart)
	activeEnd := minTime(sub.End, periodEnd)

	activeDays := 0
	if !activeStart.After(activeEnd) {
		activeDays, err = pol.CountBusinessDays(activeStart, activeEnd)
		if err != nil {
			return money.Money{}, err
		}
	}

	// 3. Define Weights: [ActiveDays, InactiveDays]
	inactiveDays := totalDays - activeDays
	weights := []int{activeDays, inactiveDays}

	// 4. Use your Allocator to split the money
	// This handles the rounding/penny logic automatically!
	parts, err := allocate.Split(sub.TotalAmount, weights)
	if err != nil {
		return money.Money{}, err
	}

	// parts[0] is the amount for ActiveDays (The Charge)
	// parts[1] is the amount for InactiveDays (The Refund/Uncharged)
	return parts[0], nil
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
