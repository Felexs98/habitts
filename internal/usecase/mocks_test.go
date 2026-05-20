package usecase

import (
	"context"
	"time"

	"habit-tracker/internal/domain"
)

type mockUserRepo struct {
	byEmail map[string]*domain.User
	byID    map[uint]*domain.User
	nextID  uint
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		byEmail: make(map[string]*domain.User),
		byID:    make(map[uint]*domain.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	m.nextID++
	user.ID = m.nextID
	cp := *user
	m.byEmail[user.Email] = &cp
	m.byID[user.ID] = &cp
	return nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	if u, ok := m.byEmail[email]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) FindByID(_ context.Context, id uint) (*domain.User, error) {
	if u, ok := m.byID[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	if _, ok := m.byID[user.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *user
	m.byID[user.ID] = &cp
	m.byEmail[user.Email] = &cp
	return nil
}

type mockHabitRepo struct {
	habits map[uint]*domain.Habit
	nextID uint
}

func newMockHabitRepo() *mockHabitRepo {
	return &mockHabitRepo{habits: make(map[uint]*domain.Habit)}
}

func (m *mockHabitRepo) Create(_ context.Context, habit *domain.Habit) error {
	m.nextID++
	habit.ID = m.nextID
	cp := *habit
	m.habits[habit.ID] = &cp
	return nil
}

func (m *mockHabitRepo) FindByID(_ context.Context, id uint) (*domain.Habit, error) {
	if h, ok := m.habits[id]; ok {
		cp := *h
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockHabitRepo) ListByUserID(_ context.Context, userID uint) ([]domain.Habit, error) {
	var list []domain.Habit
	for _, h := range m.habits {
		if h.UserID == userID {
			list = append(list, *h)
		}
	}
	return list, nil
}

func (m *mockHabitRepo) Update(_ context.Context, habit *domain.Habit) error {
	if _, ok := m.habits[habit.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *habit
	m.habits[habit.ID] = &cp
	return nil
}

func (m *mockHabitRepo) Delete(_ context.Context, id, userID uint) error {
	h, ok := m.habits[id]
	if !ok || h.UserID != userID {
		return domain.ErrNotFound
	}
	delete(m.habits, id)
	return nil
}

type mockCheckInRepo struct {
	byHabit map[uint][]domain.CheckIn
	nextID  uint
}

func newMockCheckInRepo() *mockCheckInRepo {
	return &mockCheckInRepo{byHabit: make(map[uint][]domain.CheckIn)}
}

func (m *mockCheckInRepo) Create(_ context.Context, checkIn *domain.CheckIn) error {
	m.nextID++
	checkIn.ID = m.nextID
	cp := *checkIn
	m.byHabit[checkIn.HabitID] = append(m.byHabit[checkIn.HabitID], cp)
	return nil
}

func (m *mockCheckInRepo) ListByHabitID(_ context.Context, habitID uint) ([]domain.CheckIn, error) {
	return append([]domain.CheckIn(nil), m.byHabit[habitID]...), nil
}

func (m *mockCheckInRepo) ExistsOnDate(_ context.Context, habitID uint, date time.Time) (bool, error) {
	d := truncateDate(date)
	for _, c := range m.byHabit[habitID] {
		if truncateDate(c.Date).Equal(d) {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockCheckInRepo) CountByHabitID(_ context.Context, habitID uint) (int64, error) {
	return int64(len(m.byHabit[habitID])), nil
}

func (m *mockCheckInRepo) CheckedHabitIDsOnDate(_ context.Context, userID uint, date time.Time) (map[uint]bool, error) {
	d := truncateDate(date)
	out := make(map[uint]bool)
	for habitID, list := range m.byHabit {
		for _, c := range list {
			if truncateDate(c.Date).Equal(d) {
				out[habitID] = true
				break
			}
		}
	}
	return out, nil
}
