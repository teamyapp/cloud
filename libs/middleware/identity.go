package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const gRPCAuthorizationKey = "Authorization"

func ServerHTTPWithIdentity(
	dataCollector telemetry.DataCollector,
	identityAPIEndpoint string,
) Middleware[http.HandlerFunc] {
	return withIdentity(dataCollector, identityAPIEndpoint, func(request *http.Request) (string, error) {
		ct := request.Context()
		value := request.Header.Get("Authorization")
		if len(value) == 0 {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "authorization header not found",
			}
			dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return "", internalErr
		}

		parts := strings.Split(value, " ")
		if len(parts) != 2 {
			internalErr := &errs.Error{
				Code:    errs.InvalidFormat,
				Message: fmt.Sprintf("authotization header must have 2 parts: header=%v", value),
			}
			dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return "", internalErr
		}

		if parts[0] != "Bearer" {
			internalErr := &errs.Error{
				Code:    errs.InvalidFormat,
				Message: fmt.Sprintf("missing beginning Bearer: header=%v", value),
			}
			dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return "", internalErr
		}

		return parts[1], nil
	})
}

func ServerWebSocketWithIdentity(
	dataCollector telemetry.DataCollector,
	identityAPIEndpoint string,
) Middleware[http.HandlerFunc] {
	return withIdentity(dataCollector, identityAPIEndpoint, func(request *http.Request) (string, error) {
		return request.URL.Query().Get("accessToken"), nil
	})
}

func ServerGRPCWithIdentity(dataCollector telemetry.DataCollector, identityAPIEndpoint string) grpc.UnaryServerInterceptor {
	verifyTokenURL := fmt.Sprintf("%s/verify-token", identityAPIEndpoint)
	return func(ct context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ct)
		if !ok {
			return handler(ct, req)
		}

		values := md.Get(gRPCAuthorizationKey)
		if len(values) > 0 {
			accessToken := values[0]
			updatedCt, err := ctxWithUserID(dataCollector, ct, verifyTokenURL, accessToken)
			if err != nil {
				dataCollector.Logger.LogWithContext(ct, telemetry.Warning, telemetry.Props{telemetry.CauseProp: err})
			} else {
				ct = updatedCt
			}
		}

		return handler(ct, req)
	}
}

func ClientGRPCWithIdentity(getAccessToken func() string) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if getAccessToken != nil {
			accessToken := getAccessToken()
			if len(accessToken) > 0 {
				ct = metadata.AppendToOutgoingContext(ct, gRPCAuthorizationKey, accessToken)
			}
		}

		return invoker(ct, method, req, reply, cc, opts...)
	}
}

func withIdentity(
	dataCollector telemetry.DataCollector,
	identityAPIEndpoint string,
	getBearerToken func(request *http.Request) (string, error),
) Middleware[http.HandlerFunc] {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		verifyTokenURL := fmt.Sprintf("%s/verify-token", identityAPIEndpoint)
		return func(writer http.ResponseWriter, request *http.Request) {
			ct := request.Context()
			token, err := getBearerToken(request)
			if err != nil {
				dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			} else {
				updatedCt, err := ctxWithUserID(dataCollector, request.Context(), verifyTokenURL, token)
				if err != nil {
					dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
				} else {
					request = request.WithContext(updatedCt)
				}
			}

			handlerFunc(writer, request)
		}
	}
}

func ctxWithUserID(dataCollector telemetry.DataCollector, ct context.Context, verifyTokenURL string, accessToken string) (context.Context, *errs.Error) {
	res, err := http.Post(
		verifyTokenURL,
		"text/plain",
		bytes.NewReader([]byte(accessToken)))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	internalErr := errs.GetFromHTTPErr(res)
	if internalErr != nil {
		dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})

		return nil, internalErr
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	userID, err := strconv.ParseUint(string(buf), 10, 64)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.InvalidFormat,
			EmbedErr: err,
		}
		dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return ctx.NewContextWithUserID(ct, userID), nil
}
