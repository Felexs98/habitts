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

func (s *HabitService) achievementIDsToGrant(
	ctx context.Context,
	userID uint,
	habit *domain.Habit,
	user *domain.User,
	totalCheckInsAfter int64,
) ([]uint, []domain.Achievement, error) {
	var ids []uint
	var granted []domain.Achievement

	if err := s.tryCollectGrant(ctx, userID, domain.AchievementCodeFirstCheckIn, totalCheckInsAfter == 1, &ids, &granted); err != nil {
		return nil, nil, err
	}
	if err := s.tryCollectGrant(ctx, userID, domain.AchievementCodeCheckIns10, totalCheckInsAfter >= 10, &ids, &granted); err != nil {
		return nil, nil, err
	}
	if err := s.tryCollectGrant(ctx, userID, domain.AchievementCodeLevel3, user.Level >= 3, &ids, &granted); err != nil {
		return nil, nil, err
	}
	if err := s.tryCollectGrant(ctx, userID, domain.AchievementCodeStreak7, habit.MaxStreak >= 7, &ids, &granted); err != nil {
		return nil, nil, err
	}

	return ids, granted, nil
}

func (s *HabitService) tryCollectGrant(
	ctx context.Context,
	userID uint,
	code string,
	condition bool,
	ids *[]uint,
	granted *[]domain.Achievement,
) error {
	if !condition {
		return nil
	}
	ach, err := s.achievements.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	has, err := s.achievements.HasUserAchievement(ctx, userID, ach.ID)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	*ids = append(*ids, ach.ID)
	*granted = append(*granted, *ach)
	return nil
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

