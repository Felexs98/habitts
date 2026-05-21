package domain

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;size:255;not null"`
	DisplayName  *string   `json:"display_name,omitempty" gorm:"size:255"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;size:255;not null"`
	XP           int       `json:"xp" gorm:"not null;default:0"`
	Level        int       `json:"level" gorm:"not null;default:1"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
