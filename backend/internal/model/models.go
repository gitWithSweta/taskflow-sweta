package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type UserPublic struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type Project struct {
	ID          uuid.UUID
	Name        string
	Description *string
	OwnerID     uuid.UUID
	CreatedAt   time.Time
}

type Task struct {
	ID          uuid.UUID
	Title       string
	Description *string
	Status      string
	Priority    string
	ProjectID   uuid.UUID
	AssigneeID  *uuid.UUID
	CreatorID   uuid.UUID
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskPatch struct {
	Title        *string
	Description  *string
	Status       *string
	Priority     *string
	DueDate      *time.Time
	DueDateClear bool
	AssigneeID   *uuid.UUID
	AssigneeSet  bool
}
