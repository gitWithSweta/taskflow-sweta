package middleware

import (
	"net/http"

	"taskflow/internal/trace"
)

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc, ok := trace.FromTraceparent(r.Header.Get("traceparent"))
		if !ok {
			tc = trace.New()
		}
		w.Header().Set("traceparent", tc.Traceparent())
		next.ServeHTTP(w, r.WithContext(trace.IntoContext(r.Context(), tc)))
	})
}
