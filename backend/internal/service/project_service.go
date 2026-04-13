package service

import (
	"context"

	"taskflow/internal/errs"
	"taskflow/internal/model"

	"github.com/google/uuid"
)

type ProjectService struct {
	projects projectRepository
	tasks    taskRepository
	users    userRepository
}

func NewProjectService(projects projectRepository, tasks taskRepository, users userRepository) *ProjectService {
	return &ProjectService{projects: projects, tasks: tasks, users: users}
}

func (s *ProjectService) List(ctx context.Context, uid uuid.UUID, limit, offset int) ([]model.Project, int, error) {
	return s.projects.ListAccessible(ctx, uid, limit, offset)
}

func (s *ProjectService) Create(ctx context.Context, uid uuid.UUID, name string, description *string) (*model.Project, error) {
	return s.projects.Create(ctx, name, description, uid)
}

func (s *ProjectService) GetWithTasks(ctx context.Context, uid, projectID uuid.UUID) (*model.Project, []model.Task, error) {
	okAccess, err := s.projects.UserHasAccess(ctx, uid, projectID)
	if err != nil {
		return nil, nil, err
	}
	if !okAccess {
		return nil, nil, errs.ErrNotFound
	}
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	if p == nil {
		return nil, nil, errs.ErrNotFound
	}
	tasks, err := s.tasks.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	return p, tasks, nil
}

func (s *ProjectService) Patch(ctx context.Context, uid, id uuid.UUID, name *string, description *string, newOwnerID *uuid.UUID) (*model.Project, error) {
	if newOwnerID != nil {
		u, err := s.users.GetByID(ctx, *newOwnerID)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, &errs.ValidationError{Fields: map[string]string{"owner_id": "user not found"}}
		}
	}
	p, err := s.projects.Update(ctx, id, uid, name, description, newOwnerID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errs.ErrNotFound
	}
	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, uid, id uuid.UUID) error {
	return s.projects.Delete(ctx, id, uid)
}

func (s *ProjectService) Collaborators(ctx context.Context, uid, projectID uuid.UUID) ([]model.UserPublic, error) {
	okAccess, err := s.projects.UserHasAccess(ctx, uid, projectID)
	if err != nil {
		return nil, err
	}
	if !okAccess {
		return nil, errs.ErrNotFound
	}
	return s.projects.ListCollaborators(ctx, projectID)
}

func (s *ProjectService) Stats(ctx context.Context, uid, projectID uuid.UUID) (map[string]int, map[string]int, error) {
	okAccess, err := s.projects.UserHasAccess(ctx, uid, projectID)
	if err != nil {
		return nil, nil, err
	}
	if !okAccess {
		return nil, nil, errs.ErrNotFound
	}
	return s.tasks.Stats(ctx, projectID)
}
