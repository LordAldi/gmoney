package calendar

import (
	"errors"
	"time"
)

var ErrPeriodTooLong = errors.New("calculation period exceeds maximum limit of 5 years")

// Policy defines how we count time.
type Policy struct {
	Weekends map[time.Weekday]struct{}
	// Holidays stored as YYYYMMDD integers (Simple, fast map lookup)
	Holidays map[int]struct{}
}

func NewStandardPolicy() *Policy {
	return &Policy{
		Weekends: map[time.Weekday]struct{}{
			time.Saturday: {},
			time.Sunday:   {},
		},
		Holidays: make(map[int]struct{}),
	}
}

func toKey(t time.Time) int {
	y, m, d := t.Date()
	return y*10000 + int(m)*100 + d
}

func (p *Policy) AddHoliday(date time.Time) {
	p.Holidays[toKey(date)] = struct{}{}
}

// CountBusinessDays counts working days using a clear iterative loop.
// Optimization: Uses integer keys to avoid memory allocation.
// Safety: Caps iteration to prevent CPU hanging on bad inputs.
func (p *Policy) CountBusinessDays(start, end time.Time) (int, error) {
	current := truncateToDay(start)
	final := truncateToDay(end)

	if current.After(final) {
		return 0, nil
	}

	// Safety Check: Prevent massive loops (e.g., User inputs year 3000)
	// A 5-year billing cycle is a reasonable upper bound for this domain.
	if final.Sub(current).Hours() > 24*365*5 {
		return 0, ErrPeriodTooLong
	}

	count := 0

	// Pre-calculate final Unix for fast comparison
	finalUnix := final.Unix()

	for current.Unix() <= finalUnix {
		if p.isWorkingDay(current) {
			count++
		}
		// Clear, standard standard library usage.
		// Easy for any dev to understand.
		current = current.AddDate(0, 0, 1)
	}

	return count, nil
}

func (p *Policy) isWorkingDay(t time.Time) bool {
	if _, isWeekend := p.Weekends[t.Weekday()]; isWeekend {
		return false
	}
	if _, isHoliday := p.Holidays[toKey(t)]; isHoliday {
		return false
	}
	return true
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
