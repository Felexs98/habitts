package usecase

import (
	"context"
	"strings"
	"time"

	"habit-tracker/internal/domain"
)

type HabitService struct {
	habits       domain.HabitRepository
	checkIns     domain.CheckInRepository
	users        domain.UserRepository
	achievements domain.AchievementRepository
}

func NewHabitService(
	habits domain.HabitRepository,
	checkIns domain.CheckInRepository,
	users domain.UserRepository,
	achievements domain.AchievementRepository,
) *HabitService {
	return &HabitService{
		habits:       habits,
		checkIns:     checkIns,
		users:        users,
		achievements: achievements,
	}
}

type CreateHabitInput struct {
	Title       string
	Description string
	Frequency   string
	Weekday     *int
}

type UpdateHabitInput struct {
	Title       string
	Description string
	Frequency   string
	Weekday     *int
}

type ListHabitsInput struct {
	Today    bool 
	Archived bool 
}

type HabitStats struct {
	TotalCheckIns int64 `json:"total_check_ins"`
	CurrentStreak int   `json:"current_streak"`
	MaxStreak     int   `json:"max_streak"`
}

type CheckInResult struct {
	CheckIn          domain.CheckIn       `json:"check_in"`
	XPAwarded        int                  `json:"xp_awarded"`
	TotalXP          int                  `json:"total_xp"`
	Level            int                  `json:"level"`
	CurrentStreak    int                  `json:"current_streak"`
	MaxStreak        int                  `json:"max_streak"`
	NewAchievements  []domain.Achievement `json:"new_achievements,omitempty"`
}

func (s *HabitService) Create(ctx context.Context, userID uint, in CreateHabitInput) (*domain.Habit, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, domain.ErrInvalidInput
	}

	freq, err := domain.ParseFrequency(in.Frequency)
	if err != nil {
		return nil, err
	}

	weekday := in.Weekday
	if freq == domain.FrequencyWeekly && weekday == nil {
		w := int(time.Now().UTC().Weekday())
		weekday = &w
	}
	if err := domain.ValidateWeekday(freq, weekday); err != nil {
		return nil, err
	}
	if freq == domain.FrequencyDaily {
		weekday = nil
	}

	habit := &domain.Habit{
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		Frequency:   freq,
		Weekday:     weekday,
	}
	if err := s.habits.Create(ctx, habit); err != nil {
		return nil, err
	}
	return habit, nil
}

func (s *HabitService) List(ctx context.Context, userID uint, in ListHabitsInput) ([]domain.Habit, error) {
	if in.Today && in.Archived {
		return nil, domain.ErrInvalidInput
	}

	habits, err := s.habits.ListByUserID(ctx, userID, in.Archived)
	if err != nil {
		return nil, err
	}
	if !in.Today {
		return habits, nil
	}

	today := truncateDate(time.Now().UTC())
	checked, err := s.checkIns.CheckedHabitIDsOnDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}

	var due []domain.Habit
	for _, h := range habits {
		if !domain.IsDueOnDate(&h, today) {
			continue
		}
		if checked[h.ID] {
			continue
		}
		due = append(due, h)
	}
	return due, nil
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

	freq := habit.Frequency
	if strings.TrimSpace(in.Frequency) != "" {
		var err error
		freq, err = domain.ParseFrequency(in.Frequency)
		if err != nil {
			return nil, err
		}
	}
	weekday := in.Weekday
	if freq == domain.FrequencyWeekly && weekday == nil {
		weekday = habit.Weekday
	}
	if err := domain.ValidateWeekday(freq, weekday); err != nil {
		return nil, err
	}
	if freq == domain.FrequencyDaily {
		weekday = nil
	}

	habit.Title = title
	habit.Description = strings.TrimSpace(in.Description)
	habit.Frequency = freq
	habit.Weekday = weekday
	if err := s.habits.Update(ctx, habit); err != nil {
		return nil, err
	}
	return habit, nil
}

func (s *HabitService) Delete(ctx context.Context, userID, habitID uint) error {
	if _, err := s.Get(ctx, userID, habitID); err != nil {
		return err
	}
	return s.habits.Delete(ctx, habitID, userID)
}

func (s *HabitService) Restore(ctx context.Context, userID, habitID uint) error {
	if err := s.habits.Restore(ctx, habitID, userID); err != nil {
		return err
	}
	return nil
}

func (s *HabitService) CheckIn(ctx context.Context, userID, habitID uint, date time.Time) (*CheckInResult, error) {
	habit, err := s.Get(ctx, userID, habitID)
	if err != nil {
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

	list, err := s.checkIns.ListByHabitID(ctx, habitID, nil, nil)
	if err != nil {
		return nil, err
	}

	checkIn := &domain.CheckIn{
		HabitID: habitID,
		Date:    d,
	}
	listWithNew := append(list, *checkIn)
	streak := domain.CalcStreakEndingOn(listWithNew, d)
	habit.CurrentStreak = streak
	if streak > habit.MaxStreak {
		habit.MaxStreak = streak
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.XP += domain.XPPerCheckIn
	user.Level = domain.LevelFromXP(user.XP)

	totalBefore, err := s.totalUserCheckIns(ctx, userID)
	if err != nil {
		return nil, err
	}
	achIDs, newAchievements, err := s.achievementIDsToGrant(ctx, userID, habit, user, totalBefore+1)
	if err != nil {
		return nil, err
	}

	if err := s.checkIns.Apply(ctx, &domain.CheckInApply{
		CheckIn:               checkIn,
		Habit:                 habit,
		User:                  user,
		AchievementIDsToGrant: achIDs,
	}); err != nil {
		return nil, err
	}

	return &CheckInResult{
		CheckIn:         *checkIn,
		XPAwarded:       domain.XPPerCheckIn,
		TotalXP:         user.XP,
		Level:           user.Level,
		CurrentStreak:   habit.CurrentStreak,
		MaxStreak:       habit.MaxStreak,
		NewAchievements: newAchievements,
	}, nil
}

func (s *HabitService) Stats(ctx context.Context, userID, habitID uint) (*HabitStats, error) {
	habit, err := s.Get(ctx, userID, habitID)
	if err != nil {
		return nil, err
	}

	total, err := s.checkIns.CountByHabitID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	return &HabitStats{
		TotalCheckIns: total,
		CurrentStreak: habit.CurrentStreak,
		MaxStreak:     habit.MaxStreak,
	}, nil
}

func truncateDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func calcStreak(checkIns []domain.CheckIn) int {
	return domain.CalcStreakEndingOn(checkIns, truncateDate(time.Now().UTC()))
}
