package seed

import (
	"context"
	"log/slog"
	"os"

	"taskflow/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustApplyIfNeeded(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config, log *slog.Logger) {
	if err := ApplyIfNeeded(ctx, pool, cfg, log); err != nil {
		log.Error("csv seed failed", "err", err)
		os.Exit(1)
	}
}
