package domain

import (
	"testing"
	"time"
)

func TestParseFrequency(t *testing.T) {
	f, err := ParseFrequency("daily")
	if err != nil || f != FrequencyDaily {
		t.Fatalf("daily: %v %q", err, f)
	}
	f, err = ParseFrequency("weekly")
	if err != nil || f != FrequencyWeekly {
		t.Fatalf("weekly: %v %q", err, f)
	}
	_, err = ParseFrequency("monthly")
	if err != ErrInvalidInput {
		t.Fatalf("monthly: got %v", err)
	}
}

func TestIsDueOnDate(t *testing.T) {
	tuesday := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC) // Tuesday
	wd := 2

	daily := &Habit{Frequency: FrequencyDaily}
	if !IsDueOnDate(daily, tuesday) {
		t.Error("daily should be due any day")
	}

	weeklyMatch := &Habit{Frequency: FrequencyWeekly, Weekday: &wd}
	if !IsDueOnDate(weeklyMatch, tuesday) {
		t.Error("weekly should be due on matching weekday")
	}

	wdMon := 1
	weeklyOther := &Habit{Frequency: FrequencyWeekly, Weekday: &wdMon}
	if IsDueOnDate(weeklyOther, tuesday) {
		t.Error("weekly should not be due on other weekday")
	}
}
