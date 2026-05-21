package usecase

import (
	"context"
	"testing"
	"time"

	"habit-tracker/internal/domain"
)

func TestHabitService_ListCheckIns_DateRange(t *testing.T) {
	svc, _, habits, checkIns := newTestHabitService()
	ctx := context.Background()
	_ = habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Read", Frequency: domain.FrequencyDaily})

	d1 := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	_ = checkIns.Create(ctx, &domain.CheckIn{HabitID: 1, Date: d1})
	_ = checkIns.Create(ctx, &domain.CheckIn{HabitID: 1, Date: d2})
	_ = checkIns.Create(ctx, &domain.CheckIn{HabitID: 1, Date: d3})

	from := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	list, err := svc.ListCheckIns(ctx, 1, 1, ListCheckInsInput{From: &from, To: &to})
	if err != nil {
		t.Fatalf("ListCheckIns() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if !truncateDate(list[0].Date).Equal(d2) {
		t.Errorf("date = %v, want %v", list[0].Date, d2)
	}
}
