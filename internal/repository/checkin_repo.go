package repository

import (
	"context"
	"errors"
	"time"

	"habit-tracker/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type checkInRepo struct {
	db *gorm.DB
}

func NewCheckInRepository(db *gorm.DB) domain.CheckInRepository {
	return &checkInRepo{db: db}
}

func (r *checkInRepo) Create(ctx context.Context, checkIn *domain.CheckIn) error {
	err := r.db.WithContext(ctx).Create(checkIn).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrDuplicateCheck
	}
	return err
}

// Apply выполняет check-in в одной транзакции: лог, streak, XP, user_achievements.
func (r *checkInRepo) Apply(ctx context.Context, apply *domain.CheckInApply) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(apply.CheckIn).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domain.ErrDuplicateCheck
			}
			return err
		}

		if err := tx.Model(&domain.Habit{}).Where("id = ?", apply.Habit.ID).
			Updates(map[string]interface{}{
				"current_streak": apply.Habit.CurrentStreak,
				"max_streak":     apply.Habit.MaxStreak,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.User{}).Where("id = ?", apply.User.ID).
			Updates(map[string]interface{}{
				"xp":    apply.User.XP,
				"level": apply.User.Level,
			}).Error; err != nil {
			return err
		}

		for _, achievementID := range apply.AchievementIDsToGrant {
			ua := &domain.UserAchievement{
				UserID:        apply.User.ID,
				AchievementID: achievementID,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(ua).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *checkInRepo) ListByHabitID(ctx context.Context, habitID uint, from, to *time.Time) ([]domain.CheckIn, error) {
	var list []domain.CheckIn
	q := r.db.WithContext(ctx).Where("habit_id = ?", habitID)
	if from != nil {
		q = q.Where("date >= ?", truncateDate(*from))
	}
	if to != nil {
		q = q.Where("date <= ?", truncateDate(*to))
	}
	err := q.Order("date DESC").Find(&list).Error
	return list, err
}

func (r *checkInRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.CheckIn{}).
		Joins("JOIN habits ON habits.id = check_ins.habit_id").
		Where("habits.user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *checkInRepo) ExistsOnDate(ctx context.Context, habitID uint, date time.Time) (bool, error) {
	var count int64
	d := truncateDate(date)
	err := r.db.WithContext(ctx).Model(&domain.CheckIn{}).
		Where("habit_id = ? AND date = ?", habitID, d).
		Count(&count).Error
	return count > 0, err
}

func (r *checkInRepo) CountByHabitID(ctx context.Context, habitID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.CheckIn{}).
		Where("habit_id = ?", habitID).
		Count(&count).Error
	return count, err
}

func (r *checkInRepo) CheckedHabitIDsOnDate(ctx context.Context, userID uint, date time.Time) (map[uint]bool, error) {
	d := truncateDate(date)
	var habitIDs []uint
	err := r.db.WithContext(ctx).
		Model(&domain.CheckIn{}).
		Joins("JOIN habits ON habits.id = check_ins.habit_id").
		Where("habits.user_id = ? AND check_ins.date = ?", userID, d).
		Distinct("check_ins.habit_id").
		Pluck("check_ins.habit_id", &habitIDs).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uint]bool, len(habitIDs))
	for _, id := range habitIDs {
		out[id] = true
	}
	return out, nil
}

func truncateDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
