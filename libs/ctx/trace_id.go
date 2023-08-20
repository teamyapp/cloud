package ctx

import (
	"context"
)

func GetTraceID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(traceIDKey).(string)
	return value, ok
}

func NewContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}
