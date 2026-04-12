package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/internal/config"
	appmigrations "taskflow/migrations"
)

func MustConnect(ctx context.Context, cfg *config.Config, log *slog.Logger) *pgxpool.Pool {
	pool, err := ConnectPool(ctx, cfg, log)
	if err != nil {
		log.Error("database connect failed", "err", err)
		os.Exit(1)
	}
	return pool
}

func ConnectPool(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	if err := runMigrations(cfg.DB.URL, log); err != nil {
		return nil, err
	}
	return openPool(ctx, cfg, log)
}

func runMigrations(databaseURL string, log *slog.Logger) error {
	d, err := iofs.New(appmigrations.Files, ".")
	if err != nil {
		return fmt.Errorf("migrations iofs: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return fmt.Errorf("migrations init: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		_, _ = m.Close()
		return fmt.Errorf("migrations up: %w", err)
	}
	sErr, dErr := m.Close()
	if err := errors.Join(sErr, dErr); err != nil {
		return fmt.Errorf("migrations close: %w", err)
	}
	log.Info("migrations ok")
	return nil
}

func openPool(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	p := cfg.DB.Pool

	pgCfg, err := pgxpool.ParseConfig(cfg.DB.URL)
	if err != nil {
		return nil, fmt.Errorf("pgx parse config: %w", err)
	}

	pgCfg.MaxConns = p.MaxConns
	pgCfg.MinConns = p.MinConns

	pgCfg.MaxConnLifetime = p.MaxConnLifetime.Duration
	pgCfg.MaxConnLifetimeJitter = p.MaxConnLifetimeJitter.Duration

	pgCfg.MaxConnIdleTime = p.MaxConnIdleTime.Duration

	pgCfg.HealthCheckPeriod = p.HealthCheckPeriod.Duration

	pgCfg.ConnConfig.ConnectTimeout = p.ConnectTimeout.Duration

	if p.StatementTimeout.Duration > 0 {
		if pgCfg.ConnConfig.RuntimeParams == nil {
			pgCfg.ConnConfig.RuntimeParams = make(map[string]string)
		}
		pgCfg.ConnConfig.RuntimeParams["statement_timeout"] =
			fmt.Sprintf("%d", p.StatementTimeout.Milliseconds())
	}

	pgCfg.ConnConfig.Tracer = QueryTracer{Log: log}

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	log.Info("db pool ready",
		"max_conns", p.MaxConns,
		"min_conns", p.MinConns,
		"max_conn_lifetime", p.MaxConnLifetime.Duration.String(),
		"max_conn_idle_time", p.MaxConnIdleTime.Duration.String(),
		"statement_timeout", p.StatementTimeout.Duration.String(),
	)
	return pool, nil
}
