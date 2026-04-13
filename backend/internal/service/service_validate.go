package service

import "time"

func validTaskStatus(s string) bool {
	switch s {
	case "todo", "in_progress", "done":
		return true
	default:
		return false
	}
}

func validTaskPriority(p string) bool {
	switch p {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func calendarDateUTC(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func dueDateBeforeTodayUTC(d time.Time) bool {
	today := calendarDateUTC(time.Now())
	return calendarDateUTC(d).Before(today)
}
