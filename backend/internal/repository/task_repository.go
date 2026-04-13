package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"taskflow/internal/errs"
	"taskflow/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) ListByProject(ctx context.Context, projectID uuid.UUID, status *string, assignee *uuid.UUID, limit, offset int) ([]model.Task, int, error) {
	baseWhere := `WHERE project_id = $1`
	args := []any{projectID}
	argN := 2
	if status != nil && *status != "" {
		baseWhere += fmt.Sprintf(` AND status = $%d::task_status`, argN)
		args = append(args, *status)
		argN++
	}
	if assignee != nil {
		baseWhere += fmt.Sprintf(` AND assignee_id = $%d`, argN)
		args = append(args, *assignee)
		argN++
	}

	countQ := `SELECT COUNT(*) FROM tasks ` + baseWhere
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := `
		SELECT id, title, description, status::text, priority::text, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks ` + baseWhere + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprint(argN) + ` OFFSET $` + fmt.Sprint(argN+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.CreatorID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, t)
	}
	return list, total, rows.Err()
}

func (r *TaskRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]model.Task, error) {
	tasks, _, err := r.ListByProject(ctx, projectID, nil, nil, 10000, 0)
	return tasks, err
}

func (r *TaskRepository) Create(ctx context.Context, title string, description *string, status, priority string, projectID, creatorID uuid.UUID, assigneeID *uuid.UUID, dueDate *time.Time) (*model.Task, error) {
	const q = `
		INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, creator_id, due_date)
		VALUES ($1, $2, $3::task_status, $4::task_priority, $5, $6, $7, $8)
		RETURNING id, title, description, status::text, priority::text, project_id, assignee_id, creator_id, due_date, created_at, updated_at`
	var t model.Task
	err := r.pool.QueryRow(ctx, q, title, description, status, priority, projectID, assigneeID, creatorID, dueDate).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatorID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	const q = `
		SELECT id, title, description, status::text, priority::text, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks WHERE id = $1`
	var t model.Task
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatorID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepository) UpdateAll(ctx context.Context, t *model.Task) (*model.Task, error) {
	const q = `
		UPDATE tasks SET
			title       = $2,
			description = $3,
			status      = $4::task_status,
			priority    = $5::task_priority,
			assignee_id = $6,
			due_date    = $7,
			creator_id  = $8,
			updated_at  = now()
		WHERE id = $1
		RETURNING id, title, description, status::text, priority::text, project_id, assignee_id, creator_id, due_date, created_at, updated_at`
	var out model.Task
	err := r.pool.QueryRow(ctx, q,
		t.ID, t.Title, t.Description, t.Status, t.Priority, t.AssigneeID, t.DueDate, t.CreatorID,
	).Scan(
		&out.ID, &out.Title, &out.Description, &out.Status, &out.Priority,
		&out.ProjectID, &out.AssigneeID, &out.CreatorID, &out.DueDate, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *TaskRepository) Stats(ctx context.Context, projectID uuid.UUID) (map[string]int, map[string]int, error) {
	byStatus := make(map[string]int)
	rows, err := r.pool.Query(ctx,
		`SELECT status::text, COUNT(*) FROM tasks WHERE project_id = $1 GROUP BY status`, projectID)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var s string
		var c int
		if err := rows.Scan(&s, &c); err != nil {
			rows.Close()
			return nil, nil, err
		}
		byStatus[s] = c
	}
	rows.Close()

	byAssignee := make(map[string]int)
	rows2, err := r.pool.Query(ctx,
		`SELECT COALESCE(assignee_id::text, 'unassigned'), COUNT(*) FROM tasks WHERE project_id = $1 GROUP BY assignee_id`, projectID)
	if err != nil {
		return nil, nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var key string
		var c int
		if err := rows2.Scan(&key, &c); err != nil {
			return nil, nil, err
		}
		byAssignee[key] = c
	}
	return byStatus, byAssignee, rows2.Err()
}
