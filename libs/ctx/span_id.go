package ctx

import (
	"context"
)

func GetSpanID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(spanIDKey).(string)
	return value, ok
}

func NewContextWithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}
