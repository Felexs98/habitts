package repository

import (
	"context"
	"errors"
	"time"

	"habit-tracker/internal/domain"

	"gorm.io/gorm"
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

func (r *checkInRepo) ListByHabitID(ctx context.Context, habitID uint) ([]domain.CheckIn, error) {
	var list []domain.CheckIn
	err := r.db.WithContext(ctx).
		Where("habit_id = ?", habitID).
		Order("date DESC").
		Find(&list).Error
	return list, err
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

func truncateDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
