package ctx

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

func GetRequestIDGRPC(ctx context.Context) string {
	return getValueGRPC(ctx, requestIDKey)
}

func GetRequestIDHttp(ctx context.Context, request *http.Request) string {
	return getValueHttp(ctx, request, requestIDKey)
}

func GetRequestID(ctx context.Context) (string, bool) {
	rawValue := ctx.Value(requestIDKey)
	if rawValue == nil {
		return "", false
	}

	value, ok := rawValue.(string)
	return value, ok
}

func NewContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func MetadataWithRequestID(ctx context.Context, requestID string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, string(requestIDKey), requestID)
	return ctx
}
