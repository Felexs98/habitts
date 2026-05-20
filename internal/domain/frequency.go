package domain

import (
	"strings"
	"time"
)

type Frequency string

const (
	FrequencyDaily  Frequency = "daily"
	FrequencyWeekly Frequency = "weekly"
)

func ParseFrequency(s string) (Frequency, error) {
	switch Frequency(strings.ToLower(strings.TrimSpace(s))) {
	case FrequencyDaily, "":
		return FrequencyDaily, nil
	case FrequencyWeekly:
		return FrequencyWeekly, nil
	default:
		return "", ErrInvalidInput
	}
}

// IsDueOnDate reports whether the habit should be done on the given calendar day.
func IsDueOnDate(h *Habit, date time.Time) bool {
	d := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	switch h.Frequency {
	case FrequencyDaily, "":
		return true
	case FrequencyWeekly:
		if h.Weekday == nil {
			return false
		}
		return int(d.Weekday()) == *h.Weekday
	default:
		return false
	}
}

func ValidateWeekday(freq Frequency, weekday *int) error {
	if freq != FrequencyWeekly {
		return nil
	}
	if weekday == nil {
		return ErrInvalidInput
	}
	if *weekday < 0 || *weekday > 6 {
		return ErrInvalidInput
	}
	return nil
}
