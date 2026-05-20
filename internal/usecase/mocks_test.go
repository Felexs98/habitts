package usecase

import (
	"context"
	"time"

	"habit-tracker/internal/domain"
)

func newTestHabitService() (*HabitService, *mockUserRepo, *mockHabitRepo, *mockCheckInRepo) {
	habits := newMockHabitRepo()
	checkIns := newMockCheckInRepo()
	users := newMockUserRepo()
	achievements := newMockAchievementRepo()
	return NewHabitService(habits, checkIns, users, achievements), users, habits, checkIns
}

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
	habits    map[uint]*domain.Habit
	deletedAt map[uint]bool
	nextID    uint
}

func newMockHabitRepo() *mockHabitRepo {
	return &mockHabitRepo{
		habits:    make(map[uint]*domain.Habit),
		deletedAt: make(map[uint]bool),
	}
}

func (m *mockHabitRepo) Create(_ context.Context, habit *domain.Habit) error {
	m.nextID++
	habit.ID = m.nextID
	cp := *habit
	m.habits[habit.ID] = &cp
	return nil
}

func (m *mockHabitRepo) FindByID(_ context.Context, id uint) (*domain.Habit, error) {
	if m.deletedAt[id] {
		return nil, domain.ErrNotFound
	}
	if h, ok := m.habits[id]; ok {
		cp := *h
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockHabitRepo) ListByUserID(_ context.Context, userID uint, archived bool) ([]domain.Habit, error) {
	var list []domain.Habit
	for id, h := range m.habits {
		if h.UserID != userID {
			continue
		}
		isArchived := m.deletedAt[id]
		if archived != isArchived {
			continue
		}
		list = append(list, *h)
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
	if !ok || h.UserID != userID || m.deletedAt[id] {
		return domain.ErrNotFound
	}
	m.deletedAt[id] = true
	return nil
}

func (m *mockHabitRepo) Restore(_ context.Context, id, userID uint) error {
	h, ok := m.habits[id]
	if !ok || h.UserID != userID || !m.deletedAt[id] {
		return domain.ErrNotFound
	}
	delete(m.deletedAt, id)
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

type mockAchievementRepo struct {
	byCode   map[string]*domain.Achievement
	all      []domain.Achievement
	byUser   map[uint][]domain.UserAchievement
	nextAID  uint
	nextUAID uint
}

func newMockAchievementRepo() *mockAchievementRepo {
	seed := []domain.Achievement{
		{ID: 1, Code: domain.AchievementCodeFirstCheckIn, Title: "Первый шаг"},
		{ID: 2, Code: domain.AchievementCodeCheckIns10, Title: "Десять побед"},
		{ID: 3, Code: domain.AchievementCodeLevel3, Title: "Опытный трекер"},
		{ID: 4, Code: domain.AchievementCodeStreak7, Title: "Неделя без перерывов"},
	}
	byCode := make(map[string]*domain.Achievement, len(seed))
	for i := range seed {
		byCode[seed[i].Code] = &seed[i]
	}
	return &mockAchievementRepo{
		byCode: byCode,
		all:    seed,
		byUser: make(map[uint][]domain.UserAchievement),
	}
}

func (m *mockAchievementRepo) FindByCode(_ context.Context, code string) (*domain.Achievement, error) {
	if a, ok := m.byCode[code]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockAchievementRepo) ListAll(_ context.Context) ([]domain.Achievement, error) {
	return append([]domain.Achievement(nil), m.all...), nil
}

func (m *mockAchievementRepo) Grant(_ context.Context, userID, achievementID uint) (*domain.UserAchievement, error) {
	has, _ := m.HasUserAchievement(context.Background(), userID, achievementID)
	if has {
		return nil, nil
	}
	m.nextUAID++
	var ach domain.Achievement
	for _, a := range m.all {
		if a.ID == achievementID {
			ach = a
			break
		}
	}
	ua := domain.UserAchievement{
		ID:            m.nextUAID,
		UserID:        userID,
		AchievementID: achievementID,
		EarnedAt:      time.Now().UTC(),
		Achievement:   ach,
	}
	m.byUser[userID] = append(m.byUser[userID], ua)
	return &ua, nil
}

func (m *mockAchievementRepo) HasUserAchievement(_ context.Context, userID, achievementID uint) (bool, error) {
	for _, ua := range m.byUser[userID] {
		if ua.AchievementID == achievementID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockAchievementRepo) ListByUserID(_ context.Context, userID uint) ([]domain.UserAchievement, error) {
	return append([]domain.UserAchievement(nil), m.byUser[userID]...), nil
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
