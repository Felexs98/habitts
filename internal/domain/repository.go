package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
}

type HabitRepository interface {
	Create(ctx context.Context, habit *Habit) error
	FindByID(ctx context.Context, id uint) (*Habit, error)
	ListByUserID(ctx context.Context, userID uint) ([]Habit, error)
	Update(ctx context.Context, habit *Habit) error
	Delete(ctx context.Context, id, userID uint) error
}

type CheckInRepository interface {
	Create(ctx context.Context, checkIn *CheckIn) error
	ListByHabitID(ctx context.Context, habitID uint) ([]CheckIn, error)
	ExistsOnDate(ctx context.Context, habitID uint, date time.Time) (bool, error)
	CountByHabitID(ctx context.Context, habitID uint) (int64, error)
}
