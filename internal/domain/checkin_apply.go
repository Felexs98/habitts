package domain

// CheckInApply — атомарная отметка: лог + streak + XP + выдача ачивок (одна транзакция в repository).
type CheckInApply struct {
	CheckIn               *CheckIn
	Habit                 *Habit
	User                  *User
	AchievementIDsToGrant []uint
}
