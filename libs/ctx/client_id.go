package ctx

import (
	"context"
)

func GetClientID(ctx context.Context) (uint64, bool) {
	value, ok := ctx.Value(clientIDKey).(uint64)
	return value, ok
}

func WithClientID(ctx context.Context, clientID uint64) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}
