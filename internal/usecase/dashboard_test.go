package usecase

import (
	"context"
	"testing"
	"time"

	"habit-tracker/internal/domain"
)

func TestHabitService_ListCheckIns(t *testing.T) {
	svc, users, habits, checkIns := newTestHabitService()
	ctx := context.Background()

	users.Create(ctx, &domain.User{ID: 1, Email: "u@test.com"})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Read"})
	day := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	checkIns.Create(ctx, &domain.CheckIn{HabitID: 1, Date: day})

	list, err := svc.ListCheckIns(ctx, 1, 1)
	if err != nil {
		t.Fatalf("ListCheckIns() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
}

func TestHabitService_Dashboard(t *testing.T) {
	svc, users, habits, checkIns := newTestHabitService()
	ctx := context.Background()

	users.Create(ctx, &domain.User{ID: 1, Email: "dash@test.com", XP: 50, Level: 1})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "A", Frequency: domain.FrequencyDaily, CurrentStreak: 3, MaxStreak: 5})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "B", Frequency: domain.FrequencyDaily, CurrentStreak: 0, MaxStreak: 0})
	today := truncateDate(time.Now().UTC())
	checkIns.Create(ctx, &domain.CheckIn{HabitID: 1, Date: today})

	dash, err := svc.Dashboard(ctx, 1)
	if err != nil {
		t.Fatalf("Dashboard() error = %v", err)
	}
	if dash.XP != 50 || dash.Level != 1 {
		t.Errorf("user xp/level = %d/%d", dash.XP, dash.Level)
	}
	if dash.ActiveHabits != 2 {
		t.Errorf("active_habits = %d", dash.ActiveHabits)
	}
	if dash.CheckInsToday != 1 {
		t.Errorf("check_ins_today = %d", dash.CheckInsToday)
	}
	if dash.BestCurrentStreak != 3 {
		t.Errorf("best_current_streak = %d", dash.BestCurrentStreak)
	}
	if len(dash.ActiveStreaks) != 1 {
		t.Errorf("active_streaks len = %d", len(dash.ActiveStreaks))
	}
}
