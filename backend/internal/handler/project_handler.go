package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"taskflow/internal/errs"
	"taskflow/internal/httpx"
	"taskflow/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService projectService
	log            *slog.Logger
}

func NewProjectHandler(projectService projectService, log *slog.Logger) *ProjectHandler {
	return &ProjectHandler{projectService: projectService, log: log}
}

func parsePagination(r *http.Request) (limit, offset int) {
	page := 1
	limit = 20
	if v := r.URL.Query().Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}
	offset = (page - 1) * limit
	return limit, offset
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	limit, offset := parsePagination(r)
	list, total, err := h.projectService.List(r.Context(), uid, limit, offset)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	out := make([]projectOut, 0, len(list))
	for i := range list {
		out = append(out, projectToOut(&list[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"projects": out,
		"total":    total,
		"page": struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		}{Limit: limit, Offset: offset},
	})
}

type createProjectBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	var body createProjectBody
	if !httpx.ReadJSON(w, r, &body) {
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		httpx.WriteValidation(w, map[string]string{"name": "is required"})
		return
	}
	p, err := h.projectService.Create(r.Context(), uid, name, body.Description)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, projectToOut(p))
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	p, tasks, err := h.projectService.GetWithTasks(r.Context(), uid, id)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	taskOuts := make([]taskOut, 0, len(tasks))
	for i := range tasks {
		t := taskToOut(&tasks[i])
		t.ProjectID = ""
		taskOuts = append(taskOuts, t)
	}
	httpx.WriteJSON(w, http.StatusOK, projectDetailToOut(p, taskOuts))
}

type patchProjectBody struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	OwnerID     *string `json:"owner_id"`
}

func (h *ProjectHandler) Patch(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	var body patchProjectBody
	if !httpx.ReadJSON(w, r, &body) {
		return
	}
	var newOwnerUUID *uuid.UUID
	if body.OwnerID != nil && strings.TrimSpace(*body.OwnerID) != "" {
		oid, perr := uuid.Parse(strings.TrimSpace(*body.OwnerID))
		if perr != nil {
			httpx.WriteValidation(w, map[string]string{"owner_id": "invalid uuid"})
			return
		}
		newOwnerUUID = &oid
	}
	p, err := h.projectService.Patch(r.Context(), uid, id, body.Name, body.Description, newOwnerUUID)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, projectToOut(p))
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	if err := h.projectService.Delete(r.Context(), uid, id); err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) Collaborators(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	users, err := h.projectService.Collaborators(r.Context(), uid, id)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	out := make([]userOut, 0, len(users))
	for i := range users {
		out = append(out, userFromPublic(users[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"users": out})
}

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, r, h.log, errs.ErrNotFound)
		return
	}
	byStatus, byAssignee, err := h.projectService.Stats(r.Context(), uid, id)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"by_status":   byStatus,
		"by_assignee": byAssignee,
	})
}
