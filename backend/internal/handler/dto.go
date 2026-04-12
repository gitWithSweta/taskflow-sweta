package handler

import (
	"time"

	"taskflow/internal/model"
)

type userOut struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type authResponse struct {
	Token string  `json:"token"`
	User  userOut `json:"user"`
}

type projectOut struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	OwnerID     string  `json:"owner_id"`
	CreatedAt   string  `json:"created_at"`
}

type projectDetailOut struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   string    `json:"created_at"`
	Tasks       []taskOut `json:"tasks"`
}

type taskOut struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	ProjectID   string  `json:"project_id,omitempty"`
	AssigneeID  *string `json:"assignee_id"`
	CreatorID   string  `json:"creator_id,omitempty"`
	DueDate     *string `json:"due_date"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func userFromPublic(u model.UserPublic) userOut {
	return userOut{ID: u.ID.String(), Name: u.Name, Email: u.Email}
}

func projectToOut(p *model.Project) projectOut {
	return projectOut{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     p.OwnerID.String(),
		CreatedAt:   p.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func projectDetailToOut(p *model.Project, tasks []taskOut) projectDetailOut {
	return projectDetailOut{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     p.OwnerID.String(),
		CreatedAt:   p.CreatedAt.UTC().Format(time.RFC3339Nano),
		Tasks:       tasks,
	}
}

func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format("2006-01-02")
	return &s
}

func taskToOut(t *model.Task) taskOut {
	var assignee *string
	if t.AssigneeID != nil {
		s := t.AssigneeID.String()
		assignee = &s
	}
	return taskOut{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		ProjectID:   t.ProjectID.String(),
		AssigneeID:  assignee,
		CreatorID:   t.CreatorID.String(),
		DueDate:     formatDatePtr(t.DueDate),
		CreatedAt:   t.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:   t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func parseDatePtr(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	d, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
