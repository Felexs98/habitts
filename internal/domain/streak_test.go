package domain

import (
	"testing"
	"time"
)

func TestCalcStreakEndingOn(t *testing.T) {
	end := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	checkIns := []CheckIn{
		{Date: end},
		{Date: end.AddDate(0, 0, -1)},
		{Date: end.AddDate(0, 0, -3)},
	}
	if got := CalcStreakEndingOn(checkIns, end); got != 2 {
		t.Errorf("got %d, want 2", got)
	}
}
