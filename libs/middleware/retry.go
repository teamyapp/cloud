package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
)

func ClientGRPCWithRetry(retry retry.Retry) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, err := retry.WithRetry(func() error {
			return invoker(ct, method, req, reply, cc, opts...)
		})
		return err
	}
}

type HttpClientExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpClientExecuteFunc func(req *http.Request) (*http.Response, error)

var _ HttpClientExecutor = (*HttpClientExecuteFunc)(nil)

func (h HttpClientExecuteFunc) Do(req *http.Request) (*http.Response, error) {
	return h(req)
}

func ClientHTTPWithRetry(dataCollector telemetry.DataCollector, retry retry.Retry) Middleware[HttpClientExecutor] {
	return func(httpClientExecutor HttpClientExecutor) HttpClientExecutor {
		return (HttpClientExecuteFunc)(func(request *http.Request) (*http.Response, error) {
			var res *http.Response
			_, err := retry.WithRetry(func() error {
				ct := request.Context()
				var err error
				res, err = httpClientExecutor.Do(request)
				if err != nil {

					dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
					return err
				}

				if res.StatusCode >= 400 {
					err = fmt.Errorf("http response error: statusCode=%v", res.StatusCode)
					dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
				}

				return err
			})

			return res, err
		})
	}

}
