package repository

import (
	"context"
	"time"

	"habit-tracker/internal/domain"

	"gorm.io/gorm"
)

type refreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Store(ctx context.Context, userID uint, jti string, expiresAt time.Time) error {
	rt := &domain.RefreshToken{
		UserID:    userID,
		JTI:       jti,
		ExpiresAt: expiresAt,
	}
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *refreshTokenRepo) Exists(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("jti = ? AND expires_at > ?", jti, time.Now().UTC()).
		Count(&count).Error
	return count > 0, err
}

func (r *refreshTokenRepo) Delete(ctx context.Context, jti string) error {
	return r.db.WithContext(ctx).Where("jti = ?", jti).Delete(&domain.RefreshToken{}).Error
}
