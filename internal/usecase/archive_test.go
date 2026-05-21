package usecase

import (
	"context"
	"testing"

	"habit-tracker/internal/domain"
)

func TestHabitService_PurgeExpiredArchives(t *testing.T) {
	svc, _, habits, _ := newTestHabitService()
	ctx := context.Background()

	_ = habits.Create(ctx, &domain.Habit{UserID: 1, Title: "A", Frequency: domain.FrequencyDaily})
	_ = habits.Delete(ctx, 1, 1)

	n, err := svc.PurgeExpiredArchives(ctx, 7)
	if err != nil {
		t.Fatalf("PurgeExpiredArchives() error = %v", err)
	}
	if n != 1 {
		t.Errorf("purged = %d, want 1", n)
	}
}
