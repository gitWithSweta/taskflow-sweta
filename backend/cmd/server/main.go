package main

import (
	"context"

	"taskflow/internal/config"
	"taskflow/internal/db"
	"taskflow/internal/logger"
	"taskflow/internal/seed"
	"taskflow/internal/server"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg)
	pool := db.MustConnect(context.Background(), cfg, log)
	defer pool.Close()
	seed.MustApplyIfNeeded(context.Background(), pool, cfg, log)
	srv := server.New(cfg, pool, log)
	srv.Start()
}
