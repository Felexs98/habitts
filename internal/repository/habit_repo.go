package repository

import (
	"context"
	"errors"
	"time"

	"habit-tracker/internal/domain"

	"gorm.io/gorm"
)

type habitRepo struct {
	db *gorm.DB
}

func NewHabitRepository(db *gorm.DB) domain.HabitRepository {
	return &habitRepo{db: db}
}

func (r *habitRepo) Create(ctx context.Context, habit *domain.Habit) error {
	return r.db.WithContext(ctx).Create(habit).Error
}

func (r *habitRepo) FindByID(ctx context.Context, id uint) (*domain.Habit, error) {
	var habit domain.Habit
	err := r.db.WithContext(ctx).First(&habit, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &habit, nil
}

func (r *habitRepo) ListByUserID(ctx context.Context, userID uint, archived bool) ([]domain.Habit, error) {
	var habits []domain.Habit
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if archived {
		q = q.Unscoped().Where("deleted_at IS NOT NULL")
	}
	err := q.Order("created_at DESC").Find(&habits).Error
	return habits, err
}

func (r *habitRepo) Update(ctx context.Context, habit *domain.Habit) error {
	return r.db.WithContext(ctx).Save(habit).Error
}

func (r *habitRepo) Delete(ctx context.Context, id, userID uint) error {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Habit{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *habitRepo) PurgeArchivedBefore(ctx context.Context, deletedBefore time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", deletedBefore).
		Delete(&domain.Habit{})
	return res.RowsAffected, res.Error
}

func (r *habitRepo) Restore(ctx context.Context, id, userID uint) error {
	res := r.db.WithContext(ctx).
		Unscoped().
		Model(&domain.Habit{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
