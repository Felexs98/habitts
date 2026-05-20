package usecase

import (
	"context"

	"habit-tracker/internal/domain"
)

type AchievementService struct {
	achievements domain.AchievementRepository
}

func NewAchievementService(achievements domain.AchievementRepository) *AchievementService {
	return &AchievementService{achievements: achievements}
}

type UserAchievementView struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Earned      bool   `json:"earned"`
	EarnedAt    string `json:"earned_at,omitempty"`
}

func (s *AchievementService) ListForUser(ctx context.Context, userID uint) ([]UserAchievementView, error) {
	all, err := s.achievements.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	earned, err := s.achievements.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	earnedMap := make(map[uint]domain.UserAchievement, len(earned))
	for _, ua := range earned {
		earnedMap[ua.AchievementID] = ua
	}

	out := make([]UserAchievementView, 0, len(all))
	for _, a := range all {
		v := UserAchievementView{
			Code:        a.Code,
			Title:       a.Title,
			Description: a.Description,
			Earned:      false,
		}
		if ua, ok := earnedMap[a.ID]; ok {
			v.Earned = true
			v.EarnedAt = ua.EarnedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		out = append(out, v)
	}
	return out, nil
}

func (s *HabitService) grantAchievements(ctx context.Context, userID uint, habit *domain.Habit) ([]domain.Achievement, error) {
	var granted []domain.Achievement

	if habit.MaxStreak >= 7 {
		a, err := s.grantByCode(ctx, userID, domain.AchievementCodeStreak7)
		if err != nil {
			return nil, err
		}
		if a != nil {
			granted = append(granted, *a)
		}
	}

	return granted, nil
}

func (s *HabitService) grantByCode(ctx context.Context, userID uint, code string) (*domain.Achievement, error) {
	ach, err := s.achievements.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	ua, err := s.achievements.Grant(ctx, userID, ach.ID)
	if err != nil {
		return nil, err
	}
	if ua == nil {
		return nil, nil
	}
	return &ua.Achievement, nil
}
