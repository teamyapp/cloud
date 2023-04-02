package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	ProtocolProp = "Protocol"
	StageProp    = "Stage"
	HostProp     = "Host"
	MethodProp   = "Method"
	PathProp     = "Path"
	HeadersProp  = "Headers"
	MetadataProp = "Metadata"
	BodyProp     = "Body"
	BodySizeProp = "BodySize"
)

func ServerHTTPLogRequest(dataCollector telemetry.DataCollector) Middleware[http.HandlerFunc] {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			ct := request.Context()
			buf, err := io.ReadAll(request.Body)
			if err != nil {
				internalErr := errs.NewError(errs.IO, err.Error())
				dataCollector.Logger.ErrorWithContext(ct, internalErr)
				return
			}

			requestLogProps := telemetry.Props{
				ProtocolProp: "web",
				StageProp:    "begin",
				HostProp:     request.Host,
				MethodProp:   request.Method,
				PathProp:     request.URL.Path,
				HeadersProp:  request.Header,
				BodySizeProp: len(buf),
			}
			if hasReadableBody(request.Header) {
				requestLogProps[BodyProp] = string(buf)
			}

			request.Body = io.NopCloser(bytes.NewReader(buf))
			dataCollector.Logger.LogWithContext(ct, telemetry.Info, requestLogProps)
			loggableWriter := newLoggableResponseWriter(dataCollector, writer, ct)

			// Process request
			handlerFunc(loggableWriter, request)

			responseLogProps := telemetry.Props{
				ProtocolProp: "web",
				StageProp:    "end",
				HostProp:     request.Host,
				MethodProp:   request.Method,
				PathProp:     request.URL.Path,
				HeadersProp:  writer.Header(),
				BodySizeProp: len(loggableWriter.responseBody),
			}
			if hasReadableBody(writer.Header()) {
				responseLogProps[BodyProp] = string(loggableWriter.responseBody)
			}

			dataCollector.Logger.LogWithContext(ct, telemetry.Info, responseLogProps)
		}
	}
}

func ServerGRPCLogRequest(dataCollector telemetry.DataCollector) grpc.UnaryServerInterceptor {
	return func(
		ct context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		requestBody := fmt.Sprintf("%v", req)
		requestLogProps := telemetry.Props{
			ProtocolProp: "gRPC",
			StageProp:    "begin",
			MethodProp:   info.FullMethod,
			BodySizeProp: len(requestBody),
			BodyProp:     requestBody,
		}

		md, ok := metadata.FromIncomingContext(ct)
		if ok {
			requestLogProps[MetadataProp] = fmt.Sprintf("%v", md)
		}

		dataCollector.Logger.LogWithContext(ct, telemetry.Info, requestLogProps)

		// Process request
		res, err := handler(ct, req)

		responseBody := fmt.Sprintf("%v", res)
		responseLogProps := telemetry.Props{
			ProtocolProp: "gRPC",
			StageProp:    "end",
			MethodProp:   info.FullMethod,
			BodyProp:     responseBody,
			BodySizeProp: len(responseBody),
		}
		dataCollector.Logger.LogWithContext(ct, telemetry.Info, responseLogProps)
		return res, err
	}
}

type LoggableResponseWriter struct {
	dataCollector telemetry.DataCollector
	http.ResponseWriter
	ct           context.Context
	responseBody []byte
}

var _ http.ResponseWriter = (*LoggableResponseWriter)(nil)
var _ http.Hijacker = (*LoggableResponseWriter)(nil)
var _ http.Flusher = (*LoggableResponseWriter)(nil)

func (l *LoggableResponseWriter) Write(i []byte) (int, error) {
	l.responseBody = i
	return l.ResponseWriter.Write(i)
}

func (l *LoggableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := l.ResponseWriter.(http.Hijacker)
	if !ok {
		err := errors.New("response does not implement http.Hijacker")
		l.dataCollector.Logger.ErrorWithContext(l.ct, &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		})
		return nil, nil, err
	}

	return hijacker.Hijack()
}

func (l *LoggableResponseWriter) Flush() {
	flusher, ok := l.ResponseWriter.(http.Flusher)
	if !ok {
		l.dataCollector.Logger.ErrorWithContext(l.ct, &errs.Error{
			Code:    errs.Unknown,
			Message: "response does not implement http.Flusher",
		})
		return
	}

	flusher.Flush()
}

func newLoggableResponseWriter(
	dataCollector telemetry.DataCollector,
	writer http.ResponseWriter,
	ct context.Context) *LoggableResponseWriter {
	return &LoggableResponseWriter{
		dataCollector:  dataCollector,
		ResponseWriter: writer,
		ct:             ct,
	}
}

func hasReadableBody(headers http.Header) bool {
	contentType := headers.Get("Content-Type")
	if len(contentType) == 0 {
		return false
	}

	if strings.HasPrefix(contentType, "text/") {
		return true
	}

	if strings.HasSuffix(contentType, "json") ||
		strings.HasSuffix(contentType, "xml") {
		return true
	}

	return false
}
