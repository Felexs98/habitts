package domain

import "time"

type Habit struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Frequency   Frequency `json:"frequency" gorm:"size:16;not null;default:daily"`
	Weekday     *int      `json:"weekday,omitempty" gorm:"type:smallint"` // 0=Sunday … 6=Saturday, only for weekly
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
