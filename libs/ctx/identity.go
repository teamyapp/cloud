package ctx

import (
	"context"
)

func UserIDFromContext(ctx context.Context) (uint64, bool) {
	rawValue := ctx.Value(userIDKey)
	if rawValue == nil {
		return 0, false
	}

	value, ok := rawValue.(uint64)
	return value, ok
}

func NewContextWithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
