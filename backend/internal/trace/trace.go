package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

type contextKey struct{}

type TraceContext struct {
	TraceID string
	SpanID  string
}

func New() TraceContext {
	return TraceContext{
		TraceID: randomHex(16),
		SpanID:  randomHex(8),
	}
}

func FromTraceparent(header string) (TraceContext, bool) {
	parts := strings.Split(strings.TrimSpace(header), "-")
	if len(parts) != 4 || len(parts[1]) != 32 || len(parts[2]) != 16 {
		return TraceContext{}, false
	}
	return TraceContext{
		TraceID: parts[1],
		SpanID:  randomHex(8),
	}, true
}

func (tc TraceContext) Traceparent() string {

	return "00-" + tc.TraceID + "-" + tc.SpanID + "-01"
}

func IntoContext(ctx context.Context, tc TraceContext) context.Context {
	return context.WithValue(ctx, contextKey{}, tc)
}

func FromContext(ctx context.Context) (TraceContext, bool) {
	tc, ok := ctx.Value(contextKey{}).(TraceContext)
	return tc, ok
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {

		panic("trace: crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}
