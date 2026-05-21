package usecase

import (
	"context"
	"log"
	"time"
)

// PurgeExpiredArchives удаляет из БД привычки в архиве дольше olderThanDays.
func (s *HabitService) PurgeExpiredArchives(ctx context.Context, olderThanDays int) (int64, error) {
	if olderThanDays < 1 {
		olderThanDays = 7
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -olderThanDays)
	return s.habits.PurgeArchivedBefore(ctx, cutoff)
}

// StartArchivePurgeLoop периодически чистит архив (фоновая горутина).
func StartArchivePurgeLoop(svc *HabitService, enabled bool, olderThanDays int, interval time.Duration) {
	if !enabled {
		return
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	run := func() {
		n, err := svc.PurgeExpiredArchives(context.Background(), olderThanDays)
		if err != nil {
			log.Printf("archive purge: %v", err)
			return
		}
		if n > 0 {
			log.Printf("archive purge: removed %d habit(s)", n)
		}
	}

	run()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}
