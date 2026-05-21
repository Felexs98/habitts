package domain

import "time"

type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	JTI       string    `json:"-" gorm:"column:jti;size:64;not null;uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
