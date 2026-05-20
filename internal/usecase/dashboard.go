package usecase

import (
	"context"
	"time"

	"habit-tracker/internal/domain"
)

type HabitStreakSummary struct {
	HabitID       uint   `json:"habit_id"`
	Title         string `json:"title"`
	CurrentStreak int    `json:"current_streak"`
	MaxStreak     int    `json:"max_streak"`
}

type Dashboard struct {
	Email             string               `json:"email"`
	XP                int                  `json:"xp"`
	Level             int                  `json:"level"`
	ActiveHabits      int                  `json:"active_habits"`
	HabitsDueToday    int                  `json:"habits_due_today"`
	CheckInsToday     int                  `json:"check_ins_today"`
	BestCurrentStreak int                  `json:"best_current_streak"`
	TotalCheckIns     int64                `json:"total_check_ins"`
	ActiveStreaks     []HabitStreakSummary `json:"active_streaks"`
}

func (s *HabitService) ListCheckIns(ctx context.Context, userID, habitID uint) ([]domain.CheckIn, error) {
	if _, err := s.Get(ctx, userID, habitID); err != nil {
		return nil, err
	}
	return s.checkIns.ListByHabitID(ctx, habitID)
}

func (s *HabitService) Dashboard(ctx context.Context, userID uint) (*Dashboard, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	habits, err := s.habits.ListByUserID(ctx, userID, false)
	if err != nil {
		return nil, err
	}

	today := truncateDate(time.Now().UTC())
	checked, err := s.checkIns.CheckedHabitIDsOnDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}

	dueToday, err := s.List(ctx, userID, ListHabitsInput{Today: true})
	if err != nil {
		return nil, err
	}

	dash := &Dashboard{
		Email:          user.Email,
		XP:             user.XP,
		Level:          user.Level,
		ActiveHabits:   len(habits),
		HabitsDueToday: len(dueToday),
		CheckInsToday:  len(checked),
		ActiveStreaks:  make([]HabitStreakSummary, 0),
	}

	for _, h := range habits {
		if h.CurrentStreak > dash.BestCurrentStreak {
			dash.BestCurrentStreak = h.CurrentStreak
		}
		if h.CurrentStreak > 0 {
			dash.ActiveStreaks = append(dash.ActiveStreaks, HabitStreakSummary{
				HabitID:       h.ID,
				Title:         h.Title,
				CurrentStreak: h.CurrentStreak,
				MaxStreak:     h.MaxStreak,
			})
		}
		count, err := s.checkIns.CountByHabitID(ctx, h.ID)
		if err != nil {
			return nil, err
		}
		dash.TotalCheckIns += count
	}

	return dash, nil
}
