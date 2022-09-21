package ctx

import (
	"context"
)

func UserIDFromContext(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(userIDKey).(uint64)
	return userID, ok
}

func NewContextWithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
