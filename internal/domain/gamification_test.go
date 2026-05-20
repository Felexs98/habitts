package domain

import "testing"

func TestLevelFromXP(t *testing.T) {
	tests := []struct {
		xp    int
		level int
	}{
		{xp: 0, level: 1},
		{xp: 10, level: 1},
		{xp: 99, level: 1},
		{xp: 100, level: 2},
		{xp: 199, level: 2},
		{xp: 250, level: 3},
	}

	for _, tt := range tests {
		if got := LevelFromXP(tt.xp); got != tt.level {
			t.Errorf("LevelFromXP(%d) = %d, want %d", tt.xp, got, tt.level)
		}
	}
}
