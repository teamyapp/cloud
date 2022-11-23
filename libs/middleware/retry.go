package middleware

import (
	"context"

	"github.com/teamyapp/cloud/libs/retry"
	"google.golang.org/grpc"
)

func ClientGRPCWithRetry(retry retry.Retry) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		execute := func() error {
			return invoker(ct, method, req, reply, cc, opts...)
		}
		_, err := retry.WithRetry(execute)
		return err
	}
}
