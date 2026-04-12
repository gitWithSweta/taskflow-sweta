package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"taskflow/internal/config"
	"taskflow/internal/handler"
	"taskflow/internal/repository"
	"taskflow/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg        *config.Config
	log        *slog.Logger
	pool       *pgxpool.Pool
	httpServer *http.Server
}

func New(cfg *config.Config, pool *pgxpool.Pool, log *slog.Logger) *Server {
	users := repository.NewUserRepository(pool)
	projects := repository.NewProjectRepository(pool)
	tasks := repository.NewTaskRepository(pool)
	sessions := repository.NewSessionRepository(pool)

	secret := []byte(cfg.Auth.JWTSecret)
	authSvc := service.NewAuth(users, sessions, secret, cfg.Auth.TokenTTL.Duration)
	projSvc := service.NewProject(projects, tasks, users)
	taskSvc := service.NewTask(projects, tasks, users)

	authH := handler.NewAuthHandler(authSvc, log)
	projH := handler.NewProjectHandler(projSvc, log)
	taskH := handler.NewTaskHandler(taskSvc, log)

	deps := routeDeps{
		log:      log,
		secret:   secret,
		sessions: sessions,
		authH:    authH,
		projH:    projH,
		taskH:    taskH,
	}

	s := &Server{
		cfg:  cfg,
		log:  log,
		pool: pool,
		httpServer: &http.Server{
			Addr:         cfg.HTTPAddr(),
			Handler:      newRouter(cfg, deps),
			ReadTimeout:  cfg.Server.ReadTimeout.Duration,
			WriteTimeout: cfg.Server.WriteTimeout.Duration,
			IdleTimeout:  cfg.Server.IdleTimeout.Duration,
		},
	}
	return s
}

func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

func (s *Server) Start() {
	go s.listen()
	s.waitShutdown()
}

func (s *Server) listen() {
	s.log.Info("listening", "addr", s.httpServer.Addr, "env", s.cfg.App.Env)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.Error("server", "err", err)
		os.Exit(1)
	}
}

func (s *Server) waitShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	s.log.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout.Duration)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("shutdown", "err", err)
	}
}
