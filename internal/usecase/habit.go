package usecase

import (
	"context"
	"strings"
	"time"

	"habit-tracker/internal/domain"
)

type HabitService struct {
	habits   domain.HabitRepository
	checkIns domain.CheckInRepository
}

func NewHabitService(habits domain.HabitRepository, checkIns domain.CheckInRepository) *HabitService {
	return &HabitService{habits: habits, checkIns: checkIns}
}

type CreateHabitInput struct {
	Title       string
	Description string
}

type UpdateHabitInput struct {
	Title       string
	Description string
}

type HabitStats struct {
	TotalCheckIns int64 `json:"total_check_ins"`
	CurrentStreak int   `json:"current_streak"`
}

func (s *HabitService) Create(ctx context.Context, userID uint, in CreateHabitInput) (*domain.Habit, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, domain.ErrInvalidInput
	}

	habit := &domain.Habit{
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(in.Description),
	}
	if err := s.habits.Create(ctx, habit); err != nil {
		return nil, err
	}
	return habit, nil
}

func (s *HabitService) List(ctx context.Context, userID uint) ([]domain.Habit, error) {
	return s.habits.ListByUserID(ctx, userID)
}

func (s *HabitService) Get(ctx context.Context, userID, habitID uint) (*domain.Habit, error) {
	habit, err := s.habits.FindByID(ctx, habitID)
	if err != nil {
		return nil, err
	}
	if habit.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return habit, nil
}

func (s *HabitService) Update(ctx context.Context, userID, habitID uint, in UpdateHabitInput) (*domain.Habit, error) {
	habit, err := s.Get(ctx, userID, habitID)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, domain.ErrInvalidInput
	}

	habit.Title = title
	habit.Description = strings.TrimSpace(in.Description)
	if err := s.habits.Update(ctx, habit); err != nil {
		return nil, err
	}
	return habit, nil
}

func (s *HabitService) Delete(ctx context.Context, userID, habitID uint) error {
	return s.habits.Delete(ctx, habitID, userID)
}

func (s *HabitService) CheckIn(ctx context.Context, userID, habitID uint, date time.Time) (*domain.CheckIn, error) {
	if _, err := s.Get(ctx, userID, habitID); err != nil {
		return nil, err
	}

	d := truncateDate(date)
	exists, err := s.checkIns.ExistsOnDate(ctx, habitID, d)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrDuplicateCheck
	}

	checkIn := &domain.CheckIn{
		HabitID: habitID,
		Date:    d,
	}
	if err := s.checkIns.Create(ctx, checkIn); err != nil {
		return nil, err
	}
	return checkIn, nil
}

func (s *HabitService) Stats(ctx context.Context, userID, habitID uint) (*HabitStats, error) {
	if _, err := s.Get(ctx, userID, habitID); err != nil {
		return nil, err
	}

	total, err := s.checkIns.CountByHabitID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	list, err := s.checkIns.ListByHabitID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	return &HabitStats{
		TotalCheckIns: total,
		CurrentStreak: calcStreak(list),
	}, nil
}

func truncateDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func calcStreak(checkIns []domain.CheckIn) int {
	if len(checkIns) == 0 {
		return 0
	}

	dates := make(map[string]struct{}, len(checkIns))
	for _, c := range checkIns {
		key := c.Date.Format("2006-01-02")
		dates[key] = struct{}{}
	}

	streak := 0
	day := truncateDate(time.Now().UTC())
	for {
		if _, ok := dates[day.Format("2006-01-02")]; !ok {
			break
		}
		streak++
		day = day.AddDate(0, 0, -1)
	}
	return streak
}
