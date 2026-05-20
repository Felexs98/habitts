package domain

import "time"

const AchievementCodeStreak7 = "streak_7"

type Achievement struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"uniqueIndex;size:64;not null"`
	Title       string `json:"title" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
}

type UserAchievement struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_user_achievement"`
	AchievementID uint      `json:"achievement_id" gorm:"not null;uniqueIndex:idx_user_achievement"`
	EarnedAt      time.Time `json:"earned_at"`

	User        User        `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Achievement Achievement `json:"achievement" gorm:"foreignKey:AchievementID;constraint:OnDelete:CASCADE"`
}
