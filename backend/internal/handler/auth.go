package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"taskflow/internal/errs"
	"taskflow/internal/httpx"
	"taskflow/internal/middleware"
)

type AuthHandler struct {
	svc authApplication
	log *slog.Logger
}

func NewAuthHandler(svc authApplication, log *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

type registerBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody
	if !httpx.ReadJSON(w, r, &body) {
		return
	}
	fields := map[string]string{}
	name := strings.TrimSpace(body.Name)
	email := strings.TrimSpace(strings.ToLower(body.Email))
	if name == "" {
		fields["name"] = "is required"
	}
	if email == "" {
		fields["email"] = "is required"
	} else if !validateEmail(email) {
		fields["email"] = "is invalid"
	}
	if body.Password == "" {
		fields["password"] = "is required"
	} else if len(body.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}
	if len(fields) > 0 {
		httpx.WriteValidation(w, fields)
		return
	}
	token, u, err := h.svc.Register(r.Context(), name, email, body.Password)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, authResponse{Token: token, User: userFromPublic(u)})
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if !httpx.ReadJSON(w, r, &body) {
		return
	}
	fields := map[string]string{}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	if email == "" {
		fields["email"] = "is required"
	}
	if body.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		httpx.WriteValidation(w, fields)
		return
	}
	token, u, err := h.svc.Login(r.Context(), email, body.Password)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, authResponse{Token: token, User: userFromPublic(u)})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	u, err := h.svc.Me(r.Context(), uid)
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": userFromPublic(u)})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := middleware.SessionIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	if err := h.svc.Logout(r.Context(), sessionID); err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		writeErr(w, r, h.log, errs.ErrUnauthorized)
		return
	}
	list, err := h.svc.ListUsers(r.Context())
	if err != nil {
		writeErr(w, r, h.log, err)
		return
	}
	out := make([]userOut, 0, len(list))
	for i := range list {
		out = append(out, userFromPublic(list[i]))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"users": out})
}
