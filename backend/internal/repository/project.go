package repository

import (
	"context"
	"errors"

	"taskflow/internal/errs"
	"taskflow/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) ListAccessible(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Project, int, error) {
	const countQ = `
		SELECT COUNT(DISTINCT p.id) FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
			AND (t.assignee_id = $1 OR t.creator_id = $1)
		WHERE p.owner_id = $1 OR t.id IS NOT NULL`
	var total int
	if err := r.pool.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQ = `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
			AND (t.assignee_id = $1 OR t.creator_id = $1)
		WHERE p.owner_id = $1 OR t.id IS NOT NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, listQ, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, rows.Err()
}

func (r *ProjectRepository) Create(ctx context.Context, name string, description *string, ownerID uuid.UUID) (*model.Project, error) {
	const q = `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, owner_id, created_at`
	var p model.Project
	err := r.pool.QueryRow(ctx, q, name, description, ownerID).Scan(
		&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	const q = `SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`
	var p model.Project
	err := r.pool.QueryRow(ctx, q, id).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) UserHasAccess(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM projects p WHERE p.id = $1 AND p.owner_id = $2
			UNION
			SELECT 1 FROM tasks t
			WHERE t.project_id = $1 AND (t.assignee_id = $2 OR t.creator_id = $2)
		)`
	var ok bool
	err := r.pool.QueryRow(ctx, q, projectID, userID).Scan(&ok)
	return ok, err
}

func (r *ProjectRepository) Update(ctx context.Context, projectID, callerID uuid.UUID, name *string, description *string, newOwnerID *uuid.UUID) (*model.Project, error) {
	const q = `
		UPDATE projects
		SET
			name        = COALESCE($3, name),
			description = CASE WHEN $4 THEN $5 ELSE description END,
			owner_id    = COALESCE($6::uuid, owner_id)
		WHERE id = $1 AND owner_id = $2
		RETURNING id, name, description, owner_id, created_at`
	var out model.Project
	err := r.pool.QueryRow(ctx, q,
		projectID, callerID,
		name,
		description != nil, description,
		newOwnerID,
	).Scan(&out.ID, &out.Name, &out.Description, &out.OwnerID, &out.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`, projectID).Scan(&exists)
		if !exists {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrForbidden
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ProjectRepository) ListCollaborators(ctx context.Context, projectID uuid.UUID) ([]model.UserPublic, error) {
	const q = `
		SELECT DISTINCT u.id, u.name, u.email
		FROM users u
		WHERE u.id = (SELECT owner_id FROM projects WHERE id = $1)
		   OR u.id IN (
		       SELECT assignee_id FROM tasks WHERE project_id = $1 AND assignee_id IS NOT NULL
		       UNION
		       SELECT creator_id  FROM tasks WHERE project_id = $1
		   )
		ORDER BY u.name ASC`
	rows, err := r.pool.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.UserPublic
	for rows.Next() {
		var u model.UserPublic
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *ProjectRepository) Delete(ctx context.Context, projectID, ownerID uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1 AND owner_id = $2`, projectID, ownerID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		var exists bool
		_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`, projectID).Scan(&exists)
		if !exists {
			return errs.ErrNotFound
		}
		return errs.ErrForbidden
	}
	return nil
}
