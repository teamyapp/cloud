package middleware

import (
	"context"
	"net/http"

	"github.com/teamyapp/cloud/libs/obs"
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

func ClientHTTPWithRetry(dataCollector obs.DataCollector, retry retry.Retry) func(httpRequest *http.Request) *http.Request {
	return func(preHttpRequest *http.Request) *http.Request {
		httpRequest, err := http.NewRequest(preHttpRequest.Method, preHttpRequest.Referer(), preHttpRequest.Body)
		if err != nil {
			ct := httpRequest.Context()
			dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return nil
		}

		return httpRequest
	}
}
