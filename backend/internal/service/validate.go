package service

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
