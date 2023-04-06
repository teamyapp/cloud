package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

func ServerHTTPWithRequestID(logger telemetry.Logger) Middleware[http.HandlerFunc] {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			requestID := ctx.GetRequestIDHttp(request)
			ct := request.Context()
			requestID = generateRequestIdIfNot(logger, ct, requestID)
			ct = ctx.NewContextWithRequestID(ct, requestID)
			request = request.WithContext(ct)
			handlerFunc(writer, request)
		}
	}
}

func ClientHTTPWithRequestID(logger telemetry.Logger) Middleware[web.HTTPClient] {
	return func(client web.HTTPClient) web.HTTPClient {
		return web.HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
			requestID := ctx.GetRequestIDHttp(req)
			ct := req.Context()
			requestID = generateRequestIdIfNot(logger, ct, requestID)
			ctx.SetRequestIDHttp(req, requestID)
			return client.Do(req)
		})
	}
}

func ServerGRPCWithRequestID(logger telemetry.Logger) grpc.UnaryServerInterceptor {
	return func(
		ct context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		requestID := ctx.GetRequestIDGRPC(ct)
		requestID = generateRequestIdIfNot(logger, ct, requestID)
		ct = ctx.NewContextWithRequestID(ct, requestID)
		return handler(ct, req)
	}
}

func ClientGRPCWithRequestID(logger telemetry.Logger) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		requestID := ctx.GetRequestIDGRPC(ct)
		requestID = generateRequestIdIfNot(logger, ct, requestID)
		ct = ctx.MetadataWithRequestID(ct, requestID)
		return invoker(ct, method, req, reply, cc, opts...)
	}
}

func generateRequestIdIfNot(logger telemetry.Logger, ct context.Context, requestID string) string {
	if len(requestID) == 0 {
		// it's okay to have conflicts for request ID
		randomID := uuid.New()
		requestID = randomID.String()
		logger.LogWithContext(ct, telemetry.Info, telemetry.Props{
			telemetry.RequestIDProp: requestID,
			telemetry.MessageProp:   "generate request ID",
		})
		return requestID
	}

	return requestID
}
