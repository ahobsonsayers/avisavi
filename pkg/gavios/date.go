package gavios

import "time"

// DateRange narrows flights by departure day.
// All or none of the fields can be set.
type DateRange struct {
	On     time.Time
	After  time.Time
	Before time.Time
}

// IsZero reports whether no date range is set
func (dateRange DateRange) IsZero() bool {
	return dateRange.On.IsZero() && dateRange.After.IsZero() && dateRange.Before.IsZero()
}

// InRange reports whether the given time falls within the range (by calendar day)
func (dateRange DateRange) InRange(datetime time.Time) bool {
	// If no range set, return true
	if dateRange.IsZero() {
		return true
	}

	date := datetime.Truncate(24 * time.Hour)

	// Check if falls on date
	if !dateRange.On.IsZero() {
		onDay := dateRange.On.Truncate(24 * time.Hour)
		if !date.Equal(onDay) {
			return false
		}
	}

	// Check if after date
	if !dateRange.After.IsZero() {
		afterDay := dateRange.After.Truncate(24 * time.Hour)
		if !date.After(afterDay) {
			return false
		}
	}

	// Check if before date
	if !dateRange.Before.IsZero() {
		beforeDay := dateRange.Before.Truncate(24 * time.Hour)
		if !date.Before(beforeDay) {
			return false
		}
	}

	return true
}
