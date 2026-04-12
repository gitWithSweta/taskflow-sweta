package seed

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"taskflow/internal/auth"
	"taskflow/internal/config"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const markerKey = "static_csv_demo_v1"

type projMapKey struct {
	email string
	slot  int
}

func ApplyIfNeeded(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config, log *slog.Logger) error {
	var one int
	err := pool.QueryRow(ctx, `SELECT 1 FROM app_seed_state WHERE key = $1`, markerKey).Scan(&one)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("seed state check: %w", err)
	}

	usersBytes, err := loadCSVBytes(cfg, "users.csv")
	if err != nil {
		return fmt.Errorf("load users.csv: %w", err)
	}
	projectsBytes, err := loadCSVBytes(cfg, "projects.csv")
	if err != nil {
		return fmt.Errorf("load projects.csv: %w", err)
	}
	tasksBytes, err := loadCSVBytes(cfg, "tasks.csv")
	if err != nil {
		return fmt.Errorf("load tasks.csv: %w", err)
	}

	userRows, err := parseCSV(usersBytes)
	if err != nil {
		return fmt.Errorf("parse users.csv: %w", err)
	}
	projectRows, err := parseCSV(projectsBytes)
	if err != nil {
		return fmt.Errorf("parse projects.csv: %w", err)
	}
	taskRows, err := parseCSV(tasksBytes)
	if err != nil {
		return fmt.Errorf("parse tasks.csv: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	emailToID := make(map[string]uuid.UUID)
	for i, row := range userRows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			return fmt.Errorf("users.csv row %d: need name,email,password", i+1)
		}
		name := strings.TrimSpace(row[0])
		email := strings.ToLower(strings.TrimSpace(row[1]))
		pw := strings.TrimSpace(row[2])
		if name == "" || email == "" || pw == "" {
			return fmt.Errorf("users.csv row %d: empty field", i+1)
		}
		hash, err := auth.HashPassword(pw)
		if err != nil {
			return fmt.Errorf("hash password row %d: %w", i+1, err)
		}
		var id uuid.UUID
		err = tx.QueryRow(ctx,
			`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
			name, email, hash,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("insert user %s: %w", email, err)
		}
		emailToID[email] = id
	}

	projectIDs := make(map[projMapKey]uuid.UUID)
	for i, row := range projectRows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			return fmt.Errorf("projects.csv row %d: need owner_email,slot,name,description", i+1)
		}
		ownerEmail := strings.ToLower(strings.TrimSpace(row[0]))
		slot, err := strconv.Atoi(strings.TrimSpace(row[1]))
		if err != nil || slot < 1 {
			return fmt.Errorf("projects.csv row %d: bad slot", i+1)
		}
		pname := strings.TrimSpace(row[2])
		descStr := strings.TrimSpace(row[3])
		oid, ok := emailToID[ownerEmail]
		if !ok {
			return fmt.Errorf("projects.csv row %d: unknown owner %s", i+1, ownerEmail)
		}
		var desc *string
		if descStr != "" {
			desc = &descStr
		}
		var pid uuid.UUID
		err = tx.QueryRow(ctx,
			`INSERT INTO projects (name, description, owner_id) VALUES ($1, $2, $3) RETURNING id`,
			pname, desc, oid,
		).Scan(&pid)
		if err != nil {
			return fmt.Errorf("insert project row %d: %w", i+1, err)
		}
		projectIDs[projMapKey{email: ownerEmail, slot: slot}] = pid
	}

	for i, row := range taskRows {
		if i == 0 {
			continue
		}
		if len(row) < 8 {
			return fmt.Errorf("tasks.csv row %d: need 8 columns", i+1)
		}
		ownerEmail := strings.ToLower(strings.TrimSpace(row[0]))
		slot, err := strconv.Atoi(strings.TrimSpace(row[1]))
		if err != nil || slot < 1 {
			return fmt.Errorf("tasks.csv row %d: bad project_slot", i+1)
		}
		title := strings.TrimSpace(row[2])
		descStr := strings.TrimSpace(row[3])
		status := strings.TrimSpace(row[4])
		priority := strings.TrimSpace(row[5])
		assigneeEmail := strings.ToLower(strings.TrimSpace(row[6]))
		dueStr := strings.TrimSpace(row[7])
		if title == "" {
			return fmt.Errorf("tasks.csv row %d: empty title", i+1)
		}
		if !validStatus(status) || !validPriority(priority) {
			return fmt.Errorf("tasks.csv row %d: invalid status/priority", i+1)
		}
		pid, ok := projectIDs[projMapKey{email: ownerEmail, slot: slot}]
		if !ok {
			return fmt.Errorf("tasks.csv row %d: unknown project %s slot %d", i+1, ownerEmail, slot)
		}
		creatorID, ok := emailToID[ownerEmail]
		if !ok {
			return fmt.Errorf("tasks.csv row %d: unknown owner", i+1)
		}
		var assigneeID *uuid.UUID
		if assigneeEmail != "" {
			aid, ok := emailToID[assigneeEmail]
			if !ok {
				return fmt.Errorf("tasks.csv row %d: unknown assignee %s", i+1, assigneeEmail)
			}
			assigneeID = &aid
		}
		var desc *string
		if descStr != "" {
			desc = &descStr
		}
		var due *time.Time
		if dueStr != "" {
			d, err := time.Parse("2006-01-02", dueStr)
			if err != nil {
				return fmt.Errorf("tasks.csv row %d: due_date %q: %w", i+1, dueStr, err)
			}
			due = &d
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, creator_id, due_date)
			 VALUES ($1, $2, $3::task_status, $4::task_priority, $5, $6, $7, $8)`,
			title, desc, status, priority, pid, assigneeID, creatorID, due,
		)
		if err != nil {
			return fmt.Errorf("insert task row %d: %w", i+1, err)
		}
	}

	_, err = tx.Exec(ctx, `INSERT INTO app_seed_state (key) VALUES ($1)`, markerKey)
	if err != nil {
		return fmt.Errorf("seed marker: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if log != nil {
		log.Info("csv demo seed applied", "users", len(userRows)-1, "projects", len(projectRows)-1, "tasks", len(taskRows)-1, "marker", markerKey)
	}
	return nil
}

func loadCSVBytes(cfg *config.Config, filename string) ([]byte, error) {
	if d := strings.TrimSpace(cfg.Seed.CSVDir); d != "" {
		b, err := os.ReadFile(filepath.Join(d, filename))
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	candidates := []string{
		filepath.Join("data", "seed", filename),
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "data", "seed", filename))
	}
	var lastErr error
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("read seed %q: %w (paths: %v)", filename, lastErr, candidates)
}

func parseCSV(b []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(b))
	r.TrimLeadingSpace = true
	return r.ReadAll()
}

func validStatus(s string) bool {
	switch s {
	case "todo", "in_progress", "done":
		return true
	default:
		return false
	}
}

func validPriority(p string) bool {
	switch p {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}
