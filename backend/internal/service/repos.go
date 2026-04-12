package service

import (
	"context"
	"time"

	"taskflow/internal/model"

	"github.com/google/uuid"
)

type sessionRepo interface {
	Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (uuid.UUID, error)
	Exists(ctx context.Context, sessionID uuid.UUID) (bool, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context, userID uuid.UUID) error
}

type userRepo interface {
	Create(ctx context.Context, name, email, passwordHash string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	ListAllPublic(ctx context.Context, limit int) ([]model.UserPublic, error)
}

type projectRepo interface {
	ListAccessible(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Project, int, error)
	Create(ctx context.Context, name string, description *string, ownerID uuid.UUID) (*model.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	UserHasAccess(ctx context.Context, userID, projectID uuid.UUID) (bool, error)
	Update(ctx context.Context, projectID, callerID uuid.UUID, name *string, description *string, newOwnerID *uuid.UUID) (*model.Project, error)
	Delete(ctx context.Context, projectID, ownerID uuid.UUID) error
	ListCollaborators(ctx context.Context, projectID uuid.UUID) ([]model.UserPublic, error)
}

type taskRepo interface {
	ListByProject(ctx context.Context, projectID uuid.UUID, status *string, assignee *uuid.UUID, limit, offset int) ([]model.Task, int, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]model.Task, error)
	Create(ctx context.Context, title string, description *string, status, priority string, projectID, creatorID uuid.UUID, assigneeID *uuid.UUID, dueDate *time.Time) (*model.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateAll(ctx context.Context, t *model.Task) (*model.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Stats(ctx context.Context, projectID uuid.UUID) (map[string]int, map[string]int, error)
}
