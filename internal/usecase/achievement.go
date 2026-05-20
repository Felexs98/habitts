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

func (s *HabitService) grantAchievements(ctx context.Context, userID uint, habit *domain.Habit, user *domain.User) ([]domain.Achievement, error) {
	var granted []domain.Achievement

	totalCheckIns, err := s.totalUserCheckIns(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.tryGrant(ctx, userID, domain.AchievementCodeFirstCheckIn, totalCheckIns == 1, &granted); err != nil {
		return nil, err
	}
	if err := s.tryGrant(ctx, userID, domain.AchievementCodeCheckIns10, totalCheckIns >= 10, &granted); err != nil {
		return nil, err
	}
	if err := s.tryGrant(ctx, userID, domain.AchievementCodeLevel3, user.Level >= 3, &granted); err != nil {
		return nil, err
	}
	if err := s.tryGrant(ctx, userID, domain.AchievementCodeStreak7, habit.MaxStreak >= 7, &granted); err != nil {
		return nil, err
	}

	return granted, nil
}

func (s *HabitService) totalUserCheckIns(ctx context.Context, userID uint) (int64, error) {
	habits, err := s.habits.ListByUserID(ctx, userID, false)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, h := range habits {
		n, err := s.checkIns.CountByHabitID(ctx, h.ID)
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

func (s *HabitService) tryGrant(ctx context.Context, userID uint, code string, condition bool, granted *[]domain.Achievement) error {
	if !condition {
		return nil
	}
	a, err := s.grantByCode(ctx, userID, code)
	if err != nil {
		return err
	}
	if a != nil {
		*granted = append(*granted, *a)
	}
	return nil
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
