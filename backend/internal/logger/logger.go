package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"taskflow/internal/config"
	"taskflow/internal/trace"
)

func istLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}

func New(cfg *config.Config) *slog.Logger {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.App.LogLevel.Level,
	})
	return slog.New(traceHandler{
		Handler: base,
		loc:     istLocation(),
	})
}

type traceHandler struct {
	slog.Handler
	loc *time.Location
}

func (h traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if !r.Time.IsZero() {
		r.Time = r.Time.In(h.loc)
	}
	if tc, ok := trace.FromContext(ctx); ok {
		r.AddAttrs(
			slog.String("trace_id", tc.TraceID),
			slog.String("span_id", tc.SpanID),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{Handler: h.Handler.WithAttrs(attrs), loc: h.loc}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{Handler: h.Handler.WithGroup(name), loc: h.loc}
}
