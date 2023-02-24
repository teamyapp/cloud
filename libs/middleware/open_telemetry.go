package middleware

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	grpc_codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const gRPCStatusCodeKey = attribute.Key("rpc.grpc.status_code")

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

var _ http.ResponseWriter = (*responseRecorder)(nil)
var _ http.Hijacker = (*responseRecorder)(nil)

func (r *responseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *responseRecorder) Write(bytes []byte) (int, error) {
	return r.ResponseWriter.Write(bytes)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		err := errors.New("response does not implement http.Hijacker")
		return nil, nil, err
	}

	return hijacker.Hijack()
}

func ServerHTTPWithOpenTelemetry(
	getPatternFunc GetPatternFunc,
) Middleware[http.HandlerFunc] {
	traceProvider := otel.GetTracerProvider()
	textMapProvider := otel.GetTextMapPropagator()
	tracer := traceProvider.Tracer(telemetry.TeamyOpenTelemetryTracerName)
	spanKindOpt := trace.WithSpanKind(trace.SpanKindServer)
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			spanRequestOpt := trace.WithAttributes(httpconv.ServerRequest(request.Host, request)...)
			pattern, found := getPatternFunc(request)
			if !found {
				pattern = request.URL.Path
			}

			ct := request.Context()
			ct = textMapProvider.Extract(ct, propagation.HeaderCarrier(request.Header))
			spanName := fmt.Sprintf("(HTTP)(%v)%v", request.Method, pattern)
			ct, span := tracer.Start(ct, spanName, spanKindOpt, spanRequestOpt)
			defer span.End()

			spanContext := span.SpanContext()
			ct = ctx.NewContextWithTraceID(ct, spanContext.TraceID().String())
			ct = ctx.NewContextWithSpanID(ct, spanContext.SpanID().String())

			resWriter := &responseRecorder{
				ResponseWriter: writer,
				statusCode:     http.StatusOK,
			}
			request = request.WithContext(ct)
			handlerFunc(resWriter, request)
			span.SetStatus(httpconv.ServerStatus(resWriter.statusCode))
		}
	}
}

func ClientHTTPWithOpenTelemetry(
	getPatternFunc GetPatternFunc,
) Middleware[web.HTTPClient] {
	traceProvider := otel.GetTracerProvider()
	textMapProvider := otel.GetTextMapPropagator()
	tracer := traceProvider.Tracer(telemetry.TeamyOpenTelemetryTracerName)
	spanKindOpt := trace.WithSpanKind(trace.SpanKindClient)
	return func(client web.HTTPClient) web.HTTPClient {
		return func(ct context.Context, request *http.Request) (*http.Response, error) {
			spanRequestOpt := trace.WithAttributes(httpconv.ClientRequest(request)...)
			pattern, found := getPatternFunc(request)
			if !found {
				pattern = request.URL.Path
			}

			spanName := fmt.Sprintf("(HTTP)(%v)%v", request.Method, pattern)
			ct, span := tracer.Start(ct, spanName, spanKindOpt, spanRequestOpt)
			defer span.End()

			spanContext := span.SpanContext()
			ct = ctx.NewContextWithTraceID(ct, spanContext.TraceID().String())
			ct = ctx.NewContextWithSpanID(ct, spanContext.SpanID().String())

			textMapProvider.Inject(ct, propagation.HeaderCarrier(request.Header))
			request = request.WithContext(ct)
			response, err := client.Do(ct, request)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return nil, err
			}

			span.SetAttributes(httpconv.ClientResponse(response)...)
			span.SetStatus(httpconv.ClientStatus(response.StatusCode))
			return response, nil
		}
	}
}

func ServerGRPCUnaryWithOpenTelemetry() grpc.UnaryServerInterceptor {
	traceProvider := otel.GetTracerProvider()
	textMapProvider := otel.GetTextMapPropagator()
	tracer := traceProvider.Tracer(telemetry.TeamyOpenTelemetryTracerName)
	spanKindOpt := trace.WithSpanKind(trace.SpanKindServer)
	return func(
		ct context.Context,
		request interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ct)
		if !ok {
			md = metadata.MD{}
		}

		ct = textMapProvider.Extract(ct, metadataCarrier(md))
		name := fmt.Sprintf("(gRPC)%v", info.FullMethod)
		ct, span := tracer.Start(ct, name, spanKindOpt)
		defer span.End()

		spanContext := span.SpanContext()
		ct = ctx.NewContextWithTraceID(ct, spanContext.TraceID().String())
		ct = ctx.NewContextWithSpanID(ct, spanContext.SpanID().String())

		response, err := handler(ct, request)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(gRPCStatusCodeKey.Int64(int64(s.Code())))
		} else {
			span.SetAttributes(gRPCStatusCodeKey.Int64(int64(grpc_codes.OK)))
		}

		return response, err
	}
}

func ClientGRPCUnaryWithOpenTelemetry() grpc.UnaryClientInterceptor {
	traceProvider := otel.GetTracerProvider()
	textMapProvider := otel.GetTextMapPropagator()
	tracer := traceProvider.Tracer(telemetry.TeamyOpenTelemetryTracerName)
	spanKindOpt := trace.WithSpanKind(trace.SpanKindClient)
	return func(
		ct context.Context,
		method string,
		req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		targetOps := trace.WithAttributes(targetAttr(cc.Target())...)
		name := fmt.Sprintf("(gRPC)%v", method)
		ct, span := tracer.Start(ct, name, spanKindOpt, targetOps)
		defer span.End()

		spanContext := span.SpanContext()
		ct = ctx.NewContextWithTraceID(ct, spanContext.TraceID().String())
		ct = ctx.NewContextWithSpanID(ct, spanContext.SpanID().String())

		md, ok := metadata.FromOutgoingContext(ct)
		if !ok {
			md = metadata.MD{}
		}

		textMapProvider.Inject(ct, metadataCarrier(md))
		ct = metadata.NewOutgoingContext(ct, md)
		err := invoker(ct, method, req, reply, cc, opts...)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(gRPCStatusCodeKey.Int64(int64(s.Code())))
		} else {
			span.SetAttributes(gRPCStatusCodeKey.Int64(int64(grpc_codes.OK)))
		}

		return err
	}
}

func targetAttr(target string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		return nil
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil
	}

	return []attribute.KeyValue{
		semconv.NetPeerName(host),
		semconv.NetPeerPort(portNum),
	}
}

type metadataCarrier metadata.MD

var _ propagation.TextMapCarrier = (*metadataCarrier)(nil)

func (m metadataCarrier) Get(key string) string {
	values := (metadata.MD)(m).Get(strings.ToLower(key))
	if len(values) < 1 {
		return ""
	}

	return values[0]
}

func (m metadataCarrier) Set(key string, value string) {
	(metadata.MD)(m).Set(strings.ToLower(key), value)
}

func (m metadataCarrier) Keys() []string {
	keys := make([]string, (metadata.MD)(m).Len())
	for key := range (metadata.MD)(m) {
		keys = append(keys, key)
	}

	return keys
}
