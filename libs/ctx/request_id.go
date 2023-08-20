package ctx

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

func GetRequestIDGRPC(ctx context.Context) string {
	return getValueGRPC(ctx, requestIDKey)
}

func GetRequestIDHttp(request *http.Request) string {
	return getValueHttp(request, requestIDKey)
}

func GetRequestID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(requestIDKey).(string)
	return value, ok
}

func NewContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func MetadataWithRequestID(ctx context.Context, requestID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, string(requestIDKey), requestID)
}

func SetRequestIDHttp(request *http.Request, requestID string) {
	setValueHttp(request, requestIDKey, requestID)
}
