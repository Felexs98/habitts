package domain

import "time"

type CheckIn struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	HabitID   uint      `json:"habit_id" gorm:"not null;index:idx_habit_date,unique"`
	Date      time.Time `json:"date" gorm:"type:date;not null;index:idx_habit_date,unique"`
	CreatedAt time.Time `json:"created_at"`

	Habit Habit `json:"-" gorm:"foreignKey:HabitID;constraint:OnDelete:CASCADE"`
}
