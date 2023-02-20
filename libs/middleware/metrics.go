package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

type ServerHTTPMetrics interface {
	ReportHTTPIncomingRequest(path string)
	ReportHTTPIncomingRequestResponseTime(path string, duration time.Duration)
}

type ClientHTTPMetrics interface {
	ReportHTTPOutgoingRequest(target string, path string)
	ReportHTTPOutgoingRequestResponseTime(target string, path string, duration time.Duration)
}

type ServerGRPCMetrics interface {
	ReportGRPCIncomingRequest(method string)
	ReportGRPCIncomingRequestResponseTime(method string, duration time.Duration)
}

type ClientGRPCMetrics interface {
	ReportGRPCOutgoingRequest(target string, method string)
	ReportGRPCOutgoingRequestResponseTime(target string, method string, duration time.Duration)
}

func ServerHTTPWithMetrics(metrics ServerHTTPMetrics) Middleware[http.HandlerFunc] {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			path := request.URL.Path
			metrics.ReportHTTPIncomingRequest(path)
			startAt := time.Now()

			handlerFunc(writer, request)

			responseTime := time.Now().Sub(startAt)
			metrics.ReportHTTPIncomingRequestResponseTime(path, responseTime)
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

func ClientHTTPWithMetrics(metrics ClientHTTPMetrics) Middleware[web.HTTPClient] {
	return func(client web.HTTPClient) web.HTTPClient {
		return func(ct context.Context, req *http.Request) (*http.Response, error) {
			metrics.ReportHTTPOutgoingRequest(req.URL.Host, req.URL.Path)
			startAt := time.Now()

			res, err := client.Do(ct, req)

			responseTime := time.Now().Sub(startAt)
			metrics.ReportHTTPOutgoingRequestResponseTime(req.URL.Host, req.URL.Path, responseTime)
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
