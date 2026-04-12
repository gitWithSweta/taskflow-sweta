package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (uuid.UUID, error) {
	const q = `INSERT INTO user_sessions (user_id, expires_at) VALUES ($1, $2) RETURNING id`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, q, userID, expiresAt).Scan(&id)
	return id, err
}

func (r *SessionRepository) Exists(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM user_sessions WHERE id = $1 AND expires_at > now())`
	var ok bool
	err := r.pool.QueryRow(ctx, q, sessionID).Scan(&ok)
	return ok, err
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, sessionID)
	return err
}

func (r *SessionRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_sessions WHERE user_id = $1 AND expires_at < now()`, userID)
	return err
}
