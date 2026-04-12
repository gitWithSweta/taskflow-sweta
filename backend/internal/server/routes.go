package server

import (
	"log/slog"
	"net/http"

	"taskflow/internal/config"
	"taskflow/internal/handler"
	"taskflow/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type routeDeps struct {
	log      *slog.Logger
	secret   []byte
	sessions middleware.SessionValidator // checked on every authenticated request
	authH    *handler.AuthHandler
	projH    *handler.ProjectHandler
	taskH    *handler.TaskHandler
}

func registerHealth(r chi.Router) {
	r.Get("/healthz", healthz)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func registerAPIRoutes(r chi.Router, d routeDeps) {
	r.Route("/api", func(r chi.Router) {
		r.Use(chimw.RequestID)
		r.Use(middleware.Tracing)
		r.Use(middleware.StructuredRequestLog(d.log))

		registerAuthRoutes(r, d)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(d.secret, d.sessions))
			registerProtectedAuthRoutes(r, d)
			registerUserRoutes(r, d)
			registerProjectRoutes(r, d)
			registerTaskRoutes(r, d)
		})
	})
}

func registerAuthRoutes(r chi.Router, d routeDeps) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", d.authH.Register)
		r.Post("/login", d.authH.Login)
	})
}

func registerProtectedAuthRoutes(r chi.Router, d routeDeps) {
	r.Get("/auth/me", d.authH.Me)
	r.Post("/auth/logout", d.authH.Logout)
}

func registerUserRoutes(r chi.Router, d routeDeps) {
	r.Get("/users", d.authH.ListUsers)
}

func registerProjectRoutes(r chi.Router, d routeDeps) {
	r.Get("/projects", d.projH.List)
	r.Post("/projects", d.projH.Create)
	r.Get("/projects/{id}/collaborators", d.projH.Collaborators)
	r.Get("/projects/{id}/stats", d.projH.Stats)
	r.Get("/projects/{id}/tasks", d.taskH.List)
	r.Post("/projects/{id}/tasks", d.taskH.Create)
	r.Get("/projects/{id}", d.projH.Get)
	r.Patch("/projects/{id}", d.projH.Patch)
	r.Delete("/projects/{id}", d.projH.Delete)
}

func registerTaskRoutes(r chi.Router, d routeDeps) {
	r.Patch("/tasks/{id}", d.taskH.Patch)
	r.Delete("/tasks/{id}", d.taskH.Delete)
}

func newRouter(cfg *config.Config, d routeDeps) chi.Router {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	registerHealth(r)
	registerAPIRoutes(r, d)
	return r
}
