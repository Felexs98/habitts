package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"habit-tracker/internal/domain"
)

func TestHabitService_CheckIn_AwardsXP(t *testing.T) {
	users := newMockUserRepo()
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	svc := NewHabitService(habits, checkIns, users)
	ctx := context.Background()

	users.Create(ctx, &domain.User{ID: 1, Email: "u@test.com", XP: 0, Level: 1})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Read"})

	today := time.Date(2026, 5, 20, 15, 0, 0, 0, time.UTC)
	result, err := svc.CheckIn(ctx, 1, 1, today)
	if err != nil {
		t.Fatalf("CheckIn() error = %v", err)
	}
	if result.XPAwarded != domain.XPPerCheckIn {
		t.Errorf("xp_awarded = %d, want %d", result.XPAwarded, domain.XPPerCheckIn)
	}
	if result.TotalXP != 10 || result.Level != 1 {
		t.Errorf("total_xp=%d level=%d, want 10 and 1", result.TotalXP, result.Level)
	}

	user, _ := users.FindByID(ctx, 1)
	if user.XP != 10 || user.Level != 1 {
		t.Errorf("stored user xp=%d level=%d", user.XP, user.Level)
	}
}

func TestHabitService_CheckIn_LevelUp(t *testing.T) {
	users := newMockUserRepo()
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	svc := NewHabitService(habits, checkIns, users)
	ctx := context.Background()

	users.Create(ctx, &domain.User{ID: 1, Email: "u@test.com", XP: 90, Level: 1})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Run"})

	result, err := svc.CheckIn(ctx, 1, 1, time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CheckIn() error = %v", err)
	}
	if result.TotalXP != 100 || result.Level != 2 {
		t.Errorf("total_xp=%d level=%d, want 100 and 2", result.TotalXP, result.Level)
	}
}

func TestHabitService_CheckIn_Duplicate(t *testing.T) {
	users := newMockUserRepo()
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	svc := NewHabitService(habits, checkIns, users)
	ctx := context.Background()

	users.Create(ctx, &domain.User{ID: 1, Email: "u@test.com"})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Meditate"})
	day := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)

	if _, err := svc.CheckIn(ctx, 1, 1, day); err != nil {
		t.Fatalf("first CheckIn() error = %v", err)
	}
	_, err := svc.CheckIn(ctx, 1, 1, day)
	if !errors.Is(err, domain.ErrDuplicateCheck) {
		t.Errorf("second CheckIn() error = %v, want ErrDuplicateCheck", err)
	}
}

func TestHabitService_CheckIn_Forbidden(t *testing.T) {
	users := newMockUserRepo()
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	svc := NewHabitService(habits, checkIns, users)
	ctx := context.Background()

	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Other user habit"})

	_, err := svc.CheckIn(ctx, 2, 1, time.Now().UTC())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("CheckIn() error = %v, want ErrForbidden", err)
	}
}

func TestCalcStreak(t *testing.T) {
	today := truncateDate(time.Now().UTC())

	tests := []struct {
		name     string
		checkIns []domain.CheckIn
		want     int
	}{
		{
			name:     "empty",
			checkIns: nil,
			want:     0,
		},
		{
			name: "today only",
			checkIns: []domain.CheckIn{
				{Date: today},
			},
			want: 1,
		},
		{
			name: "three consecutive days",
			checkIns: []domain.CheckIn{
				{Date: today},
				{Date: today.AddDate(0, 0, -1)},
				{Date: today.AddDate(0, 0, -2)},
			},
			want: 3,
		},
		{
			name: "gap breaks streak",
			checkIns: []domain.CheckIn{
				{Date: today},
				{Date: today.AddDate(0, 0, -2)},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcStreak(tt.checkIns); got != tt.want {
				t.Errorf("calcStreak() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHabitService_List_TodayFilter(t *testing.T) {
	users := newMockUserRepo()
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	svc := NewHabitService(habits, checkIns, users)
	ctx := context.Background()

	today := time.Now().UTC()
	todayWd := int(today.Weekday())
	otherWd := (todayWd + 1) % 7

	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Daily", Frequency: domain.FrequencyDaily})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Weekly today", Frequency: domain.FrequencyWeekly, Weekday: &todayWd})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Weekly other", Frequency: domain.FrequencyWeekly, Weekday: &otherWd})
	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Daily done", Frequency: domain.FrequencyDaily})
	checkIns.Create(ctx, &domain.CheckIn{HabitID: 4, Date: truncateDate(today)})

	list, err := svc.List(ctx, 1, ListHabitsInput{Today: true})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("today list len = %d, want 2 (daily + weekly today, not done)", len(list))
	}
}

func TestHabitService_Create_WeeklyRequiresWeekdayOrDefault(t *testing.T) {
	users := newMockUserRepo()
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	svc := NewHabitService(habits, checkIns, users)
	ctx := context.Background()

	h, err := svc.Create(ctx, 1, CreateHabitInput{
		Title:     "Gym",
		Frequency: "weekly",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if h.Frequency != domain.FrequencyWeekly || h.Weekday == nil {
		t.Fatal("expected weekly with weekday set")
	}
}

func TestHabitService_Stats_Streak(t *testing.T) {
	users := newMockUserRepo()
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	svc := NewHabitService(habits, checkIns, users)
	ctx := context.Background()

	habits.Create(ctx, &domain.Habit{UserID: 1, Title: "Water"})
	today := truncateDate(time.Now().UTC())
	checkIns.Create(ctx, &domain.CheckIn{HabitID: 1, Date: today})
	checkIns.Create(ctx, &domain.CheckIn{HabitID: 1, Date: today.AddDate(0, 0, -1)})

	stats, err := svc.Stats(ctx, 1, 1)
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.TotalCheckIns != 2 {
		t.Errorf("total_check_ins = %d, want 2", stats.TotalCheckIns)
	}
	if stats.CurrentStreak != 2 {
		t.Errorf("current_streak = %d, want 2", stats.CurrentStreak)
	}
}
