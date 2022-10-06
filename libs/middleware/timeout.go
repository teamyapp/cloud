package middleware

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

func ServerHTTPWithTimeout(duration time.Duration) HTTPServerMiddleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
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
}

func ServerGRPCWithTimeout(duration time.Duration) grpc.UnaryServerInterceptor {
	return func(ct context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ct = setDeadlineIfNot(ct, duration)
		return handler(ct, req)
	}
}

func ClientGRPCWithTimout(duration time.Duration) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ct = setDeadlineIfNot(ct, duration)
		return invoker(ct, method, req, reply, cc, opts...)
	}
}

func setDeadlineIfNot(ct context.Context, duration time.Duration) context.Context {
	if _, deadlineSet := ct.Deadline(); !deadlineSet {
		ct, _ = context.WithTimeout(ct, duration)
	}

	return ct
}
