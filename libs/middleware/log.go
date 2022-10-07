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

	"github.com/teamyapp/cloud/libs/obs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func ServerHTTPLogRequest(dataCollector obs.DataCollector) HTTPServerMiddleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			ct := request.Context()
			buf, err := io.ReadAll(request.Body)
			if err != nil {
				dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return
			}

			requestLogProps := obs.Props{
				"protocol": "web",
				"stage":    "begin",
				"host":     request.URL.Host,
				"method":   request.Method,
				"path":     request.URL.Path,
				"headers":  request.Header,
				"bodySize": len(buf),
			}
			if hasReadableBody(request.Header) {
				requestLogProps["body"] = string(buf)
			}

			request.Body = io.NopCloser(bytes.NewReader(buf))
			dataCollector.Logger.LogWithContext(ct, obs.Info, requestLogProps)
			loggableWriter := newLoggableResponseWriter(dataCollector, writer, ct)

			// Process request
			handlerFunc(loggableWriter, request)

			responseLogProps := obs.Props{
				"protocol": "web",
				"stage":    "end",
				"method":   request.Method,
				"path":     request.URL.Path,
				"headers":  writer.Header(),
				"bodySize": len(loggableWriter.responseBody),
			}
			if hasReadableBody(writer.Header()) {
				responseLogProps["body"] = string(loggableWriter.responseBody)
			}

			dataCollector.Logger.LogWithContext(ct, obs.Info, requestLogProps)
		}
	}
}

func ServerGRPCLogRequest(dataCollector obs.DataCollector) grpc.UnaryServerInterceptor {
	return func(
		ct context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		requestBody := fmt.Sprintf("%v", req)
		requestLogProps := obs.Props{
			"protocol": "gRPC",
			"stage":    "begin",
			"method":   info.FullMethod,
			"body":     requestBody,
			"bodySize": len(requestBody),
		}

		md, ok := metadata.FromIncomingContext(ct)
		if ok {
			requestLogProps["metadata"] = fmt.Sprintf("%v", md)
		}

		dataCollector.Logger.LogWithContext(ct, obs.Info, requestLogProps)

		// Process request
		res, err := handler(ct, req)

		responseBody := fmt.Sprintf("%v", res)
		responseLogProps := obs.Props{
			"protocol": "gRPC",
			"stage":    "end",
			"method":   info.FullMethod,
			"body":     responseBody,
			"bodySize": len(responseBody),
		}
		dataCollector.Logger.LogWithContext(ct, obs.Info, responseLogProps)
		return res, err
	}
}

type LoggableResponseWriter struct {
	dataCollector obs.DataCollector
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
		l.dataCollector.Logger.LogWithContext(l.ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, nil, err
	}

	return hijacker.Hijack()
}

func (l *LoggableResponseWriter) Flush() {
	flusher, ok := l.ResponseWriter.(http.Flusher)
	if !ok {
		err := "response does not implement http.Flusher"
		l.dataCollector.Logger.LogWithContext(l.ct, obs.Error, obs.Props{obs.CauseProp: err})
		return
	}

	flusher.Flush()
}

func newLoggableResponseWriter(
	dataCollector obs.DataCollector,
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
