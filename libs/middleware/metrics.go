package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

type GetPatternFunc func(request *http.Request) (string, bool)

type ServerHTTPMetrics interface {
	ReportHTTPIncomingRequest(method string, pattern string)
	ReportHTTPIncomingRequestResponseTime(method string, pattern string, duration time.Duration)
}

type ClientHTTPMetrics interface {
	ReportHTTPOutgoingRequest(target string, method string, pattern string)
	ReportHTTPOutgoingRequestResponseTime(target string, method string, pattern string, duration time.Duration)
}

type ServerGRPCMetrics interface {
	ReportGRPCIncomingRequest(method string)
	ReportGRPCIncomingRequestResponseTime(method string, duration time.Duration)
}

type ClientGRPCMetrics interface {
	ReportGRPCOutgoingRequest(target string, method string)
	ReportGRPCOutgoingRequestResponseTime(target string, method string, duration time.Duration)
}

func ServerHTTPWithMetrics(metrics ServerHTTPMetrics, getPattern GetPatternFunc) Middleware[http.HandlerFunc] {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			pattern, found := getPattern(request)

			var startAt time.Time
			if found {
				metrics.ReportHTTPIncomingRequest(request.Method, pattern)
				startAt = time.Now()
			}

			handlerFunc(writer, request)

			if found {
				responseTime := time.Now().Sub(startAt)
				metrics.ReportHTTPIncomingRequestResponseTime(request.Method, pattern, responseTime)
			}
		}
	}
}

func ServerGRPCWithMetrics(metrics ServerGRPCMetrics) grpc.UnaryServerInterceptor {
	return func(
		ct context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		metrics.ReportGRPCIncomingRequest(info.FullMethod)
		startAt := time.Now()

		resp, err = handler(ct, req)

		responseTime := time.Now().Sub(startAt)
		metrics.ReportGRPCIncomingRequestResponseTime(info.FullMethod, responseTime)
		return resp, err
	}
}

func ClientHTTPWithMetrics(metrics ClientHTTPMetrics, getPattern GetPatternFunc) Middleware[web.HTTPClient] {
	return func(client web.HTTPClient) web.HTTPClient {
		return func(ct context.Context, req *http.Request) (*http.Response, error) {
			pattern, found := getPattern(req)

			var startAt time.Time
			if found {
				metrics.ReportHTTPOutgoingRequest(req.URL.Host, req.Method, pattern)
				startAt = time.Now()
			}

			res, err := client.Do(ct, req)

			if found {
				responseTime := time.Now().Sub(startAt)
				metrics.ReportHTTPOutgoingRequestResponseTime(req.URL.Host, req.Method, pattern, responseTime)
			}

			return res, err
		}
	}
}

func ClientGRPCWithMetrics(metrics ClientGRPCMetrics) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		metrics.ReportGRPCOutgoingRequest(cc.Target(), method)
		startAt := time.Now()

		err := invoker(ct, method, req, reply, cc, opts...)

		responseTime := time.Now().Sub(startAt)
		metrics.ReportGRPCOutgoingRequestResponseTime(cc.Target(), method, responseTime)
		return err
	}
}
