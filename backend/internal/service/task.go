package service

import (
	"context"
	"strings"
	"time"

	"taskflow/internal/errs"
	"taskflow/internal/model"

	"github.com/google/uuid"
)

type Task struct {
	projects projectRepo
	tasks    taskRepo
	users    userRepo
}

func NewTask(projects projectRepo, tasks taskRepo, users userRepo) *Task {
	return &Task{projects: projects, tasks: tasks, users: users}
}

func (s *Task) List(ctx context.Context, uid, projectID uuid.UUID, status *string, assignee *uuid.UUID, limit, offset int) ([]model.Task, int, error) {
	okAccess, err := s.projects.UserHasAccess(ctx, uid, projectID)
	if err != nil {
		return nil, 0, err
	}
	if !okAccess {
		return nil, 0, errs.ErrNotFound
	}
	return s.tasks.ListByProject(ctx, projectID, status, assignee, limit, offset)
}

func (s *Task) Create(ctx context.Context, uid, projectID uuid.UUID, title string, description *string, status, priority string, assignee *uuid.UUID, dueDate *time.Time) (*model.Task, error) {
	okAccess, err := s.projects.UserHasAccess(ctx, uid, projectID)
	if err != nil {
		return nil, err
	}
	if !okAccess {
		return nil, errs.ErrNotFound
	}
	if !validTaskStatus(status) {
		return nil, &errs.ValidationError{Fields: map[string]string{"status": "invalid"}}
	}
	if !validTaskPriority(priority) {
		return nil, &errs.ValidationError{Fields: map[string]string{"priority": "invalid"}}
	}
	if assignee != nil {
		u, err := s.users.GetByID(ctx, *assignee)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, &errs.ValidationError{Fields: map[string]string{"assignee_id": "user not found"}}
		}
	}
	return s.tasks.Create(ctx, title, description, status, priority, projectID, uid, assignee, dueDate)
}

func (s *Task) Patch(ctx context.Context, uid, taskID uuid.UUID, patch model.TaskPatch) (*model.Task, error) {
	cur, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, errs.ErrNotFound
	}
	okAccess, err := s.projects.UserHasAccess(ctx, uid, cur.ProjectID)
	if err != nil {
		return nil, err
	}
	if !okAccess {
		return nil, errs.ErrNotFound
	}
	p, err := s.projects.GetByID(ctx, cur.ProjectID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errs.ErrNotFound
	}
	isOwner := p.OwnerID == uid
	isAssignee := cur.AssigneeID != nil && *cur.AssigneeID == uid
	isCreator := cur.CreatorID == uid
	if !isOwner && !isAssignee && !isCreator {
		return nil, errs.ErrForbidden
	}

	next := *cur

	if patch.Title != nil {
		next.Title = strings.TrimSpace(*patch.Title)
		if next.Title == "" {
			return nil, &errs.ValidationError{Fields: map[string]string{"title": "cannot be empty"}}
		}
	}
	if patch.Description != nil {
		next.Description = patch.Description
	}
	if patch.Status != nil {
		if !validTaskStatus(*patch.Status) {
			return nil, &errs.ValidationError{Fields: map[string]string{"status": "invalid"}}
		}
		next.Status = *patch.Status
	}
	if patch.Priority != nil {
		if !validTaskPriority(*patch.Priority) {
			return nil, &errs.ValidationError{Fields: map[string]string{"priority": "invalid"}}
		}
		next.Priority = *patch.Priority
	}
	if patch.DueDateClear {
		next.DueDate = nil
	} else if patch.DueDate != nil {
		next.DueDate = patch.DueDate
	}
	if patch.AssigneeSet {
		if patch.AssigneeID == nil {
			next.AssigneeID = nil
		} else {
			u, err := s.users.GetByID(ctx, *patch.AssigneeID)
			if err != nil {
				return nil, err
			}
			if u == nil {
				return nil, &errs.ValidationError{Fields: map[string]string{"assignee_id": "user not found"}}
			}
			next.AssigneeID = patch.AssigneeID
		}
	}
	if patch.CreatorID != nil {
		if !isOwner {
			return nil, errs.ErrForbidden
		}
		u, err := s.users.GetByID(ctx, *patch.CreatorID)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, &errs.ValidationError{Fields: map[string]string{"creator_id": "user not found"}}
		}
		next.CreatorID = *patch.CreatorID
	}

	return s.tasks.UpdateAll(ctx, &next)
}

func (s *Task) Delete(ctx context.Context, uid, taskID uuid.UUID) error {
	cur, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if cur == nil {
		return errs.ErrNotFound
	}
	p, err := s.projects.GetByID(ctx, cur.ProjectID)
	if err != nil {
		return err
	}
	if p == nil {
		return errs.ErrNotFound
	}
	if p.OwnerID != uid && cur.CreatorID != uid {
		return errs.ErrForbidden
	}
	return s.tasks.Delete(ctx, taskID)
}
