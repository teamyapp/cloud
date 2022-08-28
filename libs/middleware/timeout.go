package middleware

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

// not used, not sure if we need to use
func WithWebTimeout(
	duration time.Duration,
	handlerFunc http.HandlerFunc,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ct := request.Context()
		ct, _ = context.WithTimeout(ct, duration)
		request = request.WithContext(ct)

		done := make(chan bool)

		go func() {
			defer func() {
				done <- true
			}()

			handlerFunc(writer, request)
		}()

		select {
		case <-ct.Done():
			writer.WriteHeader(http.StatusGatewayTimeout)
		case <-done:
		}
	}
}

func ServerWithGRPCTimeout(duration time.Duration) grpc.UnaryServerInterceptor {
	return func(ct context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, deadlineSet := ct.Deadline(); !deadlineSet {
			ct, _ = context.WithTimeout(ct, duration)
		}

		return handler(ct, req)
	}
}

func ClientWithGRPCTimout(duration time.Duration) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, deadlineSet := ct.Deadline(); !deadlineSet {
			ct, _ = context.WithTimeout(ct, duration)
		}

		return invoker(ct, method, req, reply, cc, opts...)
	}
}
