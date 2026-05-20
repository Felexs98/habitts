package repository

import (
	"context"
	"errors"

	"habit-tracker/internal/domain"

	"gorm.io/gorm"
)

type achievementRepo struct {
	db *gorm.DB
}

func NewAchievementRepository(db *gorm.DB) domain.AchievementRepository {
	return &achievementRepo{db: db}
}

func (r *achievementRepo) FindByCode(ctx context.Context, code string) (*domain.Achievement, error) {
	var a domain.Achievement
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *achievementRepo) ListAll(ctx context.Context) ([]domain.Achievement, error) {
	var list []domain.Achievement
	err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *achievementRepo) Grant(ctx context.Context, userID, achievementID uint) (*domain.UserAchievement, error) {
	has, err := r.HasUserAchievement(ctx, userID, achievementID)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, nil
	}

	ua := &domain.UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
	}
	if err := r.db.WithContext(ctx).Create(ua).Error; err != nil {
		return nil, err
	}
	return r.loadUserAchievement(ctx, ua.ID)
}

func (r *achievementRepo) HasUserAchievement(ctx context.Context, userID, achievementID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.UserAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count).Error
	return count > 0, err
}

func (r *achievementRepo) ListByUserID(ctx context.Context, userID uint) ([]domain.UserAchievement, error) {
	var list []domain.UserAchievement
	err := r.db.WithContext(ctx).
		Preload("Achievement").
		Where("user_id = ?", userID).
		Order("earned_at DESC").
		Find(&list).Error
	return list, err
}

func (r *achievementRepo) loadUserAchievement(ctx context.Context, id uint) (*domain.UserAchievement, error) {
	var ua domain.UserAchievement
	err := r.db.WithContext(ctx).Preload("Achievement").First(&ua, id).Error
	if err != nil {
		return nil, err
	}
	return &ua, nil
}

func SeedAchievements(db *gorm.DB) error {
	seed := []domain.Achievement{
		{
			Code:        domain.AchievementCodeFirstCheckIn,
			Title:       "Первый шаг",
			Description: "Выполните первую отметку привычки",
		},
		{
			Code:        domain.AchievementCodeCheckIns10,
			Title:       "Десять побед",
			Description: "Совершите 10 отметок по всем привычкам",
		},
		{
			Code:        domain.AchievementCodeLevel3,
			Title:       "Опытный трекер",
			Description: "Достигните 3 уровня профиля",
		},
		{
			Code:        domain.AchievementCodeStreak7,
			Title:       "Неделя без перерывов",
			Description: "Достигните серии из 7 дней выполнения привычки",
		},
	}
	for _, a := range seed {
		var existing domain.Achievement
		err := db.Where("code = ?", a.Code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&a).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}
