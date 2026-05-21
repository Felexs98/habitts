package usecase

import (
	"context"
	"testing"
)

func TestAchievementService_ListForUser(t *testing.T) {
	achievements := newMockAchievementRepo()
	svc := NewAchievementService(achievements)
	ctx := context.Background()

	list, err := svc.ListForUser(ctx, 1)
	if err != nil {
		t.Fatalf("ListForUser() error = %v", err)
	}
	if len(list) != 8 {
		t.Fatalf("len = %d, want 8 achievements in catalog", len(list))
	}
	if list[0].Earned {
		t.Error("expected not earned yet")
	}

	_, _ = achievements.Grant(ctx, 1, 1)
	list, err = svc.ListForUser(ctx, 1)
	if err != nil {
		t.Fatalf("ListForUser() error = %v", err)
	}
	if !list[0].Earned || list[0].EarnedAt == "" {
		t.Error("expected earned with timestamp")
	}
}
