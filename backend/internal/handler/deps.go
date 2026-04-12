package handler

import (
	"context"
	"time"

	"taskflow/internal/model"

	"github.com/google/uuid"
)

type authApplication interface {
	Register(ctx context.Context, name, email, password string) (token string, user model.UserPublic, err error)
	Login(ctx context.Context, email, password string) (token string, user model.UserPublic, err error)
	Me(ctx context.Context, uid uuid.UUID) (model.UserPublic, error)
	ListUsers(ctx context.Context) ([]model.UserPublic, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
}

type projectApplication interface {
	List(ctx context.Context, uid uuid.UUID, limit, offset int) ([]model.Project, int, error)
	Create(ctx context.Context, uid uuid.UUID, name string, description *string) (*model.Project, error)
	GetWithTasks(ctx context.Context, uid, projectID uuid.UUID) (*model.Project, []model.Task, error)
	Patch(ctx context.Context, uid, id uuid.UUID, name *string, description *string, newOwnerID *uuid.UUID) (*model.Project, error)
	Delete(ctx context.Context, uid, id uuid.UUID) error
	Collaborators(ctx context.Context, uid, projectID uuid.UUID) ([]model.UserPublic, error)
	Stats(ctx context.Context, uid, projectID uuid.UUID) (map[string]int, map[string]int, error)
}

type taskApplication interface {
	List(ctx context.Context, uid, projectID uuid.UUID, status *string, assignee *uuid.UUID, limit, offset int) ([]model.Task, int, error)
	Create(ctx context.Context, uid, projectID uuid.UUID, title string, description *string, status, priority string, assignee *uuid.UUID, dueDate *time.Time) (*model.Task, error)
	Patch(ctx context.Context, uid, taskID uuid.UUID, patch model.TaskPatch) (*model.Task, error)
	Delete(ctx context.Context, uid, taskID uuid.UUID) error
}
