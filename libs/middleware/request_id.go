package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/teamyapp/cloud/libs/ctx"
	"google.golang.org/grpc"
)

func WebWithRequestID(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ct := request.Context()
		requestID := ctx.GetRequestIdHttp(ct, request)
		requestID, _ = generateRequestIdIfNot(ct, requestID)

		ct = ctx.NewContextWithRequestID(ct, requestID)
		request = request.WithContext(ct)
		handlerFunc(writer, request)
	}
}

var GRPCWithRequestID grpc.UnaryServerInterceptor = func(
	ct context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	requestID := ctx.GetRequestIdGRPC(ct)
	requestID, isGenerated := generateRequestIdIfNot(ct, requestID)

	if isGenerated {
		ct = ctx.MetadataWithRequestID(ct, requestID)
	}

	res, err := handler(ct, req)
	return res, err
}

var ClientWithGRPCRequestID grpc.UnaryClientInterceptor = func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	requestID := ctx.GetRequestIdGRPC(ct)
	requestID, _ = generateRequestIdIfNot(ct, requestID)
	ct = ctx.MetadataWithRequestID(ct, requestID)

	return invoker(ct, method, req, reply, cc, opts...)
}

func generateRequestIdIfNot(ct context.Context, requestID string) (string, bool) {
	if len(requestID) == 0 {
		// it's okay to have conflicts for request ID
		randomID := uuid.New()
		requestID := randomID.String()
		log.Printf("[GRPC] Generate request ID: requestID=%v\n", requestID)
		return requestID, true
	}

	return requestID, false
}
