package usecase

import (
	"context"
	"testing"
	"time"

	"habit-tracker/internal/domain"
)

func TestHabitService_CheckIn_GrantsFirstCheckIn(t *testing.T) {
	svc, users, habits, _ := newTestHabitService()
	ctx := context.Background()

	mustCreateUser(t, ctx, users, &domain.User{ID: 1, Email: "u@test.com"})
	mustCreateHabit(t, ctx, habits, &domain.Habit{UserID: 1, Title: "Read"})

	result, err := svc.CheckIn(ctx, 1, 1, mustDate(2026, 5, 20))
	if err != nil {
		t.Fatalf("CheckIn() error = %v", err)
	}

	found := false
	for _, a := range result.NewAchievements {
		if a.Code == domain.AchievementCodeFirstCheckIn {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected first_check_in in %v", result.NewAchievements)
	}
}

func TestHabitService_CheckIn_GrantsLevel3(t *testing.T) {
	svc, users, habits, _ := newTestHabitService()
	ctx := context.Background()

	mustCreateUser(t, ctx, users, &domain.User{ID: 1, Email: "u@test.com", XP: 190, Level: 2})
	mustCreateHabit(t, ctx, habits, &domain.Habit{UserID: 1, Title: "Run"})

	result, err := svc.CheckIn(ctx, 1, 1, mustDate(2026, 5, 20))
	if err != nil {
		t.Fatalf("CheckIn() error = %v", err)
	}

	found := false
	for _, a := range result.NewAchievements {
		if a.Code == domain.AchievementCodeLevel3 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected level_3 after XP 100, achievements: %v", result.NewAchievements)
	}
}

func mustDate(y int, m int, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}
