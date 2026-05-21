package usecase

import (
	"context"
	"errors"
	"testing"

	"habit-tracker/internal/domain"
)

func TestHabitService_Delete_SoftArchive(t *testing.T) {
	svc, _, habits, _ := newTestHabitService()
	ctx := context.Background()

	mustCreateHabit(t, ctx, habits, &domain.Habit{UserID: 1, Title: "Archive me"})

	if err := svc.Delete(ctx, 1, 1); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	active, _ := svc.List(ctx, 1, ListHabitsInput{})
	if len(active) != 0 {
		t.Fatalf("active len = %d, want 0", len(active))
	}

	archived, _ := svc.List(ctx, 1, ListHabitsInput{Archived: true})
	if len(archived) != 1 {
		t.Fatalf("archived len = %d, want 1", len(archived))
	}

	_, err := svc.Get(ctx, 1, 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() after delete = %v, want ErrNotFound", err)
	}
}

func TestHabitService_Restore(t *testing.T) {
	svc, _, habits, _ := newTestHabitService()
	ctx := context.Background()

	mustCreateHabit(t, ctx, habits, &domain.Habit{UserID: 1, Title: "Restore me"})
	_ = svc.Delete(ctx, 1, 1)

	if err := svc.Restore(ctx, 1, 1); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	active, _ := svc.List(ctx, 1, ListHabitsInput{})
	if len(active) != 1 {
		t.Fatalf("active len = %d, want 1", len(active))
	}
}
