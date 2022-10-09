package ctx

import (
	"context"
)

func GetClientID(ctx context.Context) (uint64, bool) {
	rawValue := ctx.Value(clientIDKey)
	if rawValue == nil {
		return 0, false
	}

	value, ok := rawValue.(uint64)
	return value, ok
}

func WithClientID(ctx context.Context, clientID uint64) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}
