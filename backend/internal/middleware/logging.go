package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func StructuredRequestLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if log == nil {
				next.ServeHTTP(w, r)
				return
			}
			reqID := chimw.GetReqID(r.Context())
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			log.InfoContext(r.Context(), "api_request_start",
				slog.String("event", "api_request_start"),
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.String("remote_addr", r.RemoteAddr),
			)

			defer func() {
				status := ww.Status()
				if status == 0 {
					status = http.StatusOK
				}
				log.InfoContext(r.Context(), "api_request_complete",
					slog.String("event", "api_request_complete"),
					slog.String("request_id", reqID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", status),
					slog.Int64("duration_ms", time.Since(start).Milliseconds()),
					slog.Int("bytes_written", ww.BytesWritten()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
