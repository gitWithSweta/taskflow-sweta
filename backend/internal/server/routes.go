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
	log             *slog.Logger
	secret          []byte
	sessions        middleware.SessionValidator
	authHandler     *handler.AuthHandler
	projectHandler  *handler.ProjectHandler
	taskHandler     *handler.TaskHandler
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
		r.Post("/register", d.authHandler.Register)
		r.Post("/login", d.authHandler.Login)
	})
}

func registerProtectedAuthRoutes(r chi.Router, d routeDeps) {
	r.Get("/auth/me", d.authHandler.Me)
	r.Post("/auth/logout", d.authHandler.Logout)
}

func registerUserRoutes(r chi.Router, d routeDeps) {
	r.Get("/users", d.authHandler.ListUsers)
}

func registerProjectRoutes(r chi.Router, d routeDeps) {
	r.Get("/projects", d.projectHandler.List)
	r.Post("/projects", d.projectHandler.Create)
	r.Get("/projects/{id}/collaborators", d.projectHandler.Collaborators)
	r.Get("/projects/{id}/stats", d.projectHandler.Stats)
	r.Get("/projects/{id}/tasks", d.taskHandler.List)
	r.Post("/projects/{id}/tasks", d.taskHandler.Create)
	r.Get("/projects/{id}", d.projectHandler.Get)
	r.Patch("/projects/{id}", d.projectHandler.Patch)
	r.Delete("/projects/{id}", d.projectHandler.Delete)
}

func registerTaskRoutes(r chi.Router, d routeDeps) {
	r.Patch("/tasks/{id}", d.taskHandler.Patch)
	r.Delete("/tasks/{id}", d.taskHandler.Delete)
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
