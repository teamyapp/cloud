package ctx

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

func GetRequestIdGRPC(ctx context.Context) string {
	return getValueGRPC(ctx, requestIDKey)
}

func GetRequestIdHttp(ctx context.Context, request *http.Request) string {
	return getValueHttp(ctx, request, requestIDKey)
}

func NewContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func MetadataWithRequestID(ctx context.Context, requestID string) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, string(requestIDKey), requestID)

	return ctx
}
