package domain

import (
	"time"

	"gorm.io/gorm"
)

type Habit struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index:idx_habits_user_list,priority:1;index:idx_habits_user_deleted,priority:1"`
	Title         string         `json:"title" gorm:"size:255;not null"`
	Description   string         `json:"description" gorm:"type:text"`
	Frequency     Frequency      `json:"frequency" gorm:"size:16;not null;default:daily"`
	Weekday       *int           `json:"weekday,omitempty" gorm:"type:smallint"` // 0=Sunday … 6=Saturday, only for weekly
	CurrentStreak int            `json:"current_streak" gorm:"not null;default:0"`
	MaxStreak     int            `json:"max_streak" gorm:"not null;default:0"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index:idx_habits_user_list,priority:2"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index:idx_habits_user_deleted,priority:2"`

	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
