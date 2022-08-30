package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LogWebRequest(dataCollector obs.DataCollector, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return
		}

		requestLogProps := obs.Props{
			"host":     request.URL.Host,
			"method":   request.Method,
			"path":     request.URL.Path,
			"headers":  request.Header,
			"bodySize": len(buf),
		}
		responseLogProps := obs.Props{}
		requestID := ctx.GetRequestIdHttp(request.Context(), request)
		if len(requestID) > 0 {
			requestLogProps["requestId"] = requestID
			responseLogProps["requestId"] = requestID
		}

		if hasReadableBody(request.Header) {
			requestLogProps["body"] = string(buf)
		}

		request.Body = ioutil.NopCloser(bytes.NewReader(buf))
		dataCollector.Logger.Log(obs.Info, obs.MergeProps(
			obs.Props{
				"protocol": "web",
				"stage":    "begin",
			},
			requestLogProps))
		loggableWriter := newLoggableResponseWriter(dataCollector, writer)

		// Process request
		handlerFunc(loggableWriter, request)

		responseLogProps["headers"] = writer.Header()
		responseLogProps["bodySize"] = len(loggableWriter.responseBody)
		if hasReadableBody(writer.Header()) {
			responseLogProps["body"] = string(loggableWriter.responseBody)
		}

		dataCollector.Logger.Log(obs.Info, obs.MergeProps(
			obs.Props{
				"protocol": "web",
				"stage":    "end",
			},
			responseLogProps))
	}
}

func LogGRPCRequest(dataCollector obs.DataCollector) grpc.UnaryServerInterceptor {
	return func(
		ct context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		requestLogProps := obs.Props{
			"method":  info.FullMethod,
			"request": fmt.Sprintf("%v", req),
		}
		responseLogProps := obs.Props{}

		md, ok := metadata.FromIncomingContext(ct)
		if ok {
			requestLogProps["metadata"] = fmt.Sprintf("%v", md)
			requestID := ctx.GetRequestIdGRPC(ct)
			if len(requestID) > 0 {
				requestLogProps["requestId"] = requestID
				responseLogProps["requestId"] = requestID
			}
		}

		dataCollector.Logger.Log(obs.Info, obs.MergeProps(
			obs.Props{
				"protocol": "gRPC",
				"stage":    "begin",
			},
			requestLogProps))

		// Process request
		res, err := handler(ct, req)

		responseLogProps["response"] = res
		dataCollector.Logger.Log(obs.Info, obs.MergeProps(
			obs.Props{
				"protocol": "gRPC",
				"stage":    "end",
			},
			responseLogProps))
		return res, err
	}
}

type LoggableResponseWriter struct {
	dataCollector obs.DataCollector
	http.ResponseWriter
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
		l.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, nil, err
	}

	return hijacker.Hijack()
}

func (l *LoggableResponseWriter) Flush() {
	flusher, ok := l.ResponseWriter.(http.Flusher)
	if !ok {
		err := "response does not implement http.Flusher"
		l.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return
	}

	flusher.Flush()
}

func newLoggableResponseWriter(dataCollector obs.DataCollector, writer http.ResponseWriter) *LoggableResponseWriter {
	return &LoggableResponseWriter{
		dataCollector:  dataCollector,
		ResponseWriter: writer,
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
