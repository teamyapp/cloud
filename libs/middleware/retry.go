package middleware

import (
	"context"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
)

type HttpClientExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}

var _ HttpClientExecutor = (*http.Client)(nil)

type HttpClientExecuteFunc func(req *http.Request) (*http.Response, error)

var _ HttpClientExecutor = (*HttpClientExecuteFunc)(nil)

func (h HttpClientExecuteFunc) Do(req *http.Request) (*http.Response, error) {
	return h(req)
}

func ClientHTTPWithRetry(dataCollector telemetry.DataCollector, retry retry.Retry) Middleware[HttpClientExecutor] {
	return func(httpClientExecutor HttpClientExecutor) HttpClientExecutor {
		return (HttpClientExecuteFunc)(func(request *http.Request) (*http.Response, error) {
			var res *http.Response
			_, err := retry.WithRetry(func() *errs.Error {
				ct := request.Context()
				var err error
				res, err = httpClientExecutor.Do(request)
				if err != nil {
					internalErr := &errs.Error{
						Code:     errs.Unknown,
						EmbedErr: err,
					}
					dataCollector.Logger.ErrorWithContext(ct, internalErr)
					return internalErr
				}

				internalErr := errs.GetFromHTTPErr(res)
				if internalErr != nil {
					dataCollector.Logger.ErrorWithContext(ct, internalErr)
				}

				return internalErr
			})

			return res, err.ToError()
		})
	}

}

func ClientGRPCWithRetry(retry retry.Retry) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, err := retry.WithRetry(func() *errs.Error {
			return errs.FromGRPCErr(invoker(ct, method, req, reply, cc, opts...))
		})
		return errs.ToGRPCErr(err)
	}
}
