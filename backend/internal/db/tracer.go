package db

import (
	"context"
	"log/slog"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

type queryStartTimeKey struct{}

const maxSQLLogLen = 512

type QueryTracer struct {
	Log *slog.Logger
}

func (t QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if t.Log == nil {
		return ctx
	}
	sql := strings.TrimSpace(data.SQL)
	if len(sql) > maxSQLLogLen {
		sql = sql[:maxSQLLogLen] + "..."
	}
	t.Log.InfoContext(ctx, "db_query_start",
		slog.String("event", "db_query_start"),
		slog.String("request_id", chimw.GetReqID(ctx)),
		slog.String("sql", sql),
		slog.Int("arg_count", len(data.Args)),
	)
	return context.WithValue(ctx, queryStartTimeKey{}, time.Now())
}

func (t QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if t.Log == nil {
		return
	}
	level := slog.LevelInfo
	if data.Err != nil {
		level = slog.LevelWarn
	}
	var durationMs int64
	if ts, ok := ctx.Value(queryStartTimeKey{}).(time.Time); ok {
		durationMs = time.Since(ts).Milliseconds()
	}
	attrs := []slog.Attr{
		slog.String("event", "db_query_complete"),
		slog.String("request_id", chimw.GetReqID(ctx)),
		slog.Int64("duration_ms", durationMs),
		slog.String("command_tag", data.CommandTag.String()),
	}
	if data.Err != nil {
		attrs = append(attrs, slog.String("err", data.Err.Error()))
	}
	t.Log.LogAttrs(ctx, level, "db_query_complete", attrs...)
}
