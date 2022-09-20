package ctx

import (
	"context"
	"fmt"
)

func UserIDFromContext(ctx context.Context) (uint64, error) {
	userID, ok := ctx.Value(userIDKey).(uint64)
	if !ok {
		return 0, fmt.Errorf("userID not found")
	}

	return userID, nil
}

func NewContextWithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
