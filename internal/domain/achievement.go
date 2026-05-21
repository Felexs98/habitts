package domain

import "time"

const (
	AchievementCodeStreak7        = "streak_7"
	AchievementCodeStreak30       = "streak_30"
	AchievementCodeFirstCheckIn   = "first_check_in"
	AchievementCodeCheckIns10     = "check_ins_10"
	AchievementCodeCheckIns50     = "check_ins_50"
	AchievementCodeLevel3         = "level_3"
	AchievementCodeLevel5         = "level_5"
	AchievementCodeHabits5        = "habits_5"
)

type Achievement struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"uniqueIndex;size:64;not null"`
	Title       string `json:"title" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
}

type UserAchievement struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_user_achievement;index:idx_user_achievements_user,priority:1"`
	AchievementID uint      `json:"achievement_id" gorm:"not null;uniqueIndex:idx_user_achievement"`
	EarnedAt      time.Time `json:"earned_at"`

	User        User        `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Achievement Achievement `json:"achievement" gorm:"foreignKey:AchievementID;constraint:OnDelete:CASCADE"`
}
