package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"taskflow/internal/errs"
	"taskflow/internal/httpx"
	"taskflow/internal/middleware"
	"taskflow/internal/model"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskService taskService
	log         *slog.Logger
}

func NewTaskHandler(taskService taskService, log *slog.Logger) *TaskHandler {
	return &TaskHandler{taskService: taskService, log: log}
}

func isValidStatus(s string) bool {
	switch s {
	case "todo", "in_progress", "done":
		return true
	default:
		return false
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		if !isValidStatus(s) {
			httpx.WriteValidation(w, map[string]string{"status": "invalid"})
			return
		}
		status = &s
	}
	var assignee *uuid.UUID
	if a := r.URL.Query().Get("assignee"); a != "" {
		id, err := uuid.Parse(a)
		if err != nil {
			httpx.WriteValidation(w, map[string]string{"assignee": "invalid uuid"})
			return
		}
		assignee = &id
	}
	limit, offset := parsePagination(r)
	list, total, err := h.taskService.List(r.Context(), uid, pid, status, assignee, limit, offset)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	out := make([]taskOut, 0, len(list))
	for i := range list {
		out = append(out, taskToOut(&list[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"tasks": out,
		"total": total,
	})
}

type createTaskBody struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	var body createTaskBody
	if !httpx.ReadJSON(w, r, &body) {
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		httpx.WriteValidation(w, map[string]string{"title": "is required"})
		return
	}

	status := "todo"
	if body.Status != nil {
		status = *body.Status
	}
	priority := "medium"
	if body.Priority != nil {
		priority = *body.Priority
	}

	var assignee *uuid.UUID
	if body.AssigneeID != nil && *body.AssigneeID != "" {
		id, err := uuid.Parse(*body.AssigneeID)
		if err != nil {
			httpx.WriteValidation(w, map[string]string{"assignee_id": "invalid uuid"})
			return
		}
		assignee = &id
	}
	due, err := parseDatePtr(body.DueDate)
	if err != nil {
		httpx.WriteValidation(w, map[string]string{"due_date": "use YYYY-MM-DD"})
		return
	}
	t, err := h.taskService.Create(r.Context(), uid, pid, title, body.Description, status, priority, assignee, due)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, taskToOut(t))
}

type patchTaskBody struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	Status      *string         `json:"status"`
	Priority    *string         `json:"priority"`
	DueDate     *string         `json:"due_date"`
	AssigneeID json.RawMessage `json:"assignee_id"`
}

func (h *TaskHandler) Patch(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	var body patchTaskBody
	if !httpx.ReadJSON(w, r, &body) {
		return
	}

	patch := model.TaskPatch{
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
		Priority:    body.Priority,
	}

	if body.DueDate != nil {
		if *body.DueDate == "" {
			patch.DueDateClear = true
		} else {
			d, err := parseDatePtr(body.DueDate)
			if err != nil {
				httpx.WriteValidation(w, map[string]string{"due_date": "use YYYY-MM-DD"})
				return
			}
			patch.DueDate = d
		}
	}

	if len(body.AssigneeID) > 0 {
		patch.AssigneeSet = true
		if strings.TrimSpace(string(body.AssigneeID)) != "null" {
			var s string
			if err := json.Unmarshal(body.AssigneeID, &s); err != nil {
				httpx.WriteValidation(w, map[string]string{"assignee_id": "invalid"})
				return
			}
			if s != "" {
				id, err := uuid.Parse(s)
				if err != nil {
					httpx.WriteValidation(w, map[string]string{"assignee_id": "invalid uuid"})
					return
				}
				patch.AssigneeID = &id
			}
		}
	}

	saved, err := h.taskService.Patch(r.Context(), uid, tid, patch)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, taskToOut(saved))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	if err := h.taskService.Delete(r.Context(), uid, tid); err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
