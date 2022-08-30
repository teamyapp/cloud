package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"google.golang.org/grpc"
)

func WebWithRequestID(dataCollector obs.DataCollector, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ct := request.Context()
		requestID := ctx.GetRequestIdHttp(ct, request)
		requestID, _ = generateRequestIdIfNot(dataCollector, requestID)
		ct = ctx.NewContextWithRequestID(ct, requestID)
		request = request.WithContext(ct)
		handlerFunc(writer, request)
	}
}

func GRPCWithRequestID(dataCollector obs.DataCollector) grpc.UnaryServerInterceptor {
	return func(
		ct context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		requestID := ctx.GetRequestIdGRPC(ct)
		requestID, isGenerated := generateRequestIdIfNot(dataCollector, requestID)
		if isGenerated {
			ct = ctx.MetadataWithRequestID(ct, requestID)
		}

		res, err := handler(ct, req)
		return res, err
	}
}

func ClientWithGRPCRequestID(dataCollector obs.DataCollector) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		requestID := ctx.GetRequestIdGRPC(ct)
		requestID, _ = generateRequestIdIfNot(dataCollector, requestID)
		ct = ctx.MetadataWithRequestID(ct, requestID)

		return invoker(ct, method, req, reply, cc, opts...)
	}
}

func generateRequestIdIfNot(dataCollector obs.DataCollector, requestID string) (string, bool) {
	if len(requestID) == 0 {
		// it's okay to have conflicts for request ID
		randomID := uuid.New()
		requestID = randomID.String()
		dataCollector.Logger.Log(obs.Info, obs.Props{
			obs.MessageProp: obs.Props{
				"summary":   "generate request ID",
				"requestID": requestID,
			},
		})
		return requestID, true
	}

	return requestID, false
}
