package domain

const XPPerCheckIn = 10

// LevelFromXP: уровень 1 при 0 XP, далее +1 уровень каждые 100 XP.
func LevelFromXP(xp int) int {
	return xp/100 + 1
}
