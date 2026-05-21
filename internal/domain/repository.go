package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, user *User) error
}

type HabitRepository interface {
	Create(ctx context.Context, habit *Habit) error
	FindByID(ctx context.Context, id uint) (*Habit, error)
	ListByUserID(ctx context.Context, userID uint, archived bool) ([]Habit, error)
	Update(ctx context.Context, habit *Habit) error
	Delete(ctx context.Context, id, userID uint) error
	Restore(ctx context.Context, id, userID uint) error
}

type CheckInRepository interface {
	Create(ctx context.Context, checkIn *CheckIn) error
	Apply(ctx context.Context, apply *CheckInApply) error
	ListByHabitID(ctx context.Context, habitID uint) ([]CheckIn, error)
	ExistsOnDate(ctx context.Context, habitID uint, date time.Time) (bool, error)
	CountByHabitID(ctx context.Context, habitID uint) (int64, error)
	CheckedHabitIDsOnDate(ctx context.Context, userID uint, date time.Time) (map[uint]bool, error)
}

type AchievementRepository interface {
	FindByCode(ctx context.Context, code string) (*Achievement, error)
	ListAll(ctx context.Context) ([]Achievement, error)
	Grant(ctx context.Context, userID, achievementID uint) (*UserAchievement, error)
	HasUserAchievement(ctx context.Context, userID, achievementID uint) (bool, error)
	ListByUserID(ctx context.Context, userID uint) ([]UserAchievement, error)
}
