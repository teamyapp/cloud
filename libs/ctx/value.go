package ctx

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

func getValueGRPC(ctx context.Context, key key) string {
	value, ok := ctx.Value(key).(string)
	if ok && len(value) > 0 {
		return value
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(string(key))
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func getValueHttp(ctx context.Context, request *http.Request, key key) string {
	value, ok := ctx.Value(key).(string)
	if ok && len(value) > 0 {
		return value
	}

	value = request.Header.Get(string(key))
	if len(value) == 0 {
		return ""
	}

	return value
}

func setValueHttp(request *http.Request, key key, value string) {
	request.Header.Set(string(key), value)
}
