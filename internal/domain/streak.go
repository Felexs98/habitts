package domain

import "time"

// CalcStreakEndingOn counts consecutive check-in days ending on endDate (inclusive).
func CalcStreakEndingOn(checkIns []CheckIn, endDate time.Time) int {
	if len(checkIns) == 0 {
		return 0
	}

	dates := make(map[string]struct{}, len(checkIns))
	for _, c := range checkIns {
		d := time.Date(c.Date.Year(), c.Date.Month(), c.Date.Day(), 0, 0, 0, 0, time.UTC)
		dates[d.Format("2006-01-02")] = struct{}{}
	}

	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
	streak := 0
	day := end
	for {
		if _, ok := dates[day.Format("2006-01-02")]; !ok {
			break
		}
		streak++
		day = day.AddDate(0, 0, -1)
	}
	return streak
}
