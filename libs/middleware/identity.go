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
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const gRPCAuthorizationKey = "Authorization"

type IncludeIdentityWebFunc func(request *http.Request) bool
type IncludeIdentityGRPCFunc func(info *grpc.UnaryServerInfo) bool

func ServerHTTPWithIdentity(
	dataCollector telemetry.DataCollector,
	httpClient web.Client,
	identityAPIEndpoint string,
	includeIdentity IncludeIdentityWebFunc,
) Middleware[http.HandlerFunc] {
	return withIdentity(dataCollector, httpClient, identityAPIEndpoint, func(request *http.Request) (string, *errs.Error) {
		ct := request.Context()
		value := request.Header.Get("Authorization")
		if len(value) == 0 {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "authorization header not found",
			}
			return "", internalErr
		}

		parts := strings.Split(value, " ")
		if len(parts) != 2 {
			internalErr := &errs.Error{
				Code:    errs.InvalidFormat,
				Message: fmt.Sprintf("authotization header must have 2 parts: header=%v", value),
			}
			dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return "", internalErr
		}

		if parts[0] != "Bearer" {
			internalErr := &errs.Error{
				Code:    errs.InvalidFormat,
				Message: fmt.Sprintf("missing beginning Bearer: header=%v", value),
			}
			dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return "", internalErr
		}

		return parts[1], nil
	}, includeIdentity)
}

func ServerWebSocketWithIdentity(
	dataCollector telemetry.DataCollector,
	httpClient web.Client,
	identityAPIEndpoint string,
	includeIdentity IncludeIdentityWebFunc,
) Middleware[http.HandlerFunc] {
	return withIdentity(dataCollector, httpClient, identityAPIEndpoint, func(request *http.Request) (string, *errs.Error) {
		token := request.URL.Query().Get("accessToken")
		if len(token) == 0 {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "access token not found",
			}
			return "", internalErr
		}

		return token, nil
	}, includeIdentity)
}

func ServerGRPCWithIdentity(
	dataCollector telemetry.DataCollector,
	httpClient web.Client,
	identityAPIEndpoint string,
	includeIdentity IncludeIdentityGRPCFunc,
) grpc.UnaryServerInterceptor {
	verifyTokenURL := fmt.Sprintf("%s/verify-token", identityAPIEndpoint)
	return func(ct context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if includeIdentity(info) {
			md, ok := metadata.FromIncomingContext(ct)
			if !ok {
				return handler(ct, req)
			}

			values := md.Get(gRPCAuthorizationKey)
			if len(values) > 0 {
				accessToken := values[0]
				updatedCt, err := ctxWithUserID(dataCollector, httpClient, verifyTokenURL, ct, accessToken)
				if err != nil {
					dataCollector.Logger.WarningWithContext(ct, err.String())
				} else {
					ct = updatedCt
				}
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
	httpClient web.Client,
	identityAPIEndpoint string,
	getBearerToken func(request *http.Request) (string, *errs.Error),
	includeIdentity IncludeIdentityWebFunc,
) Middleware[http.HandlerFunc] {
	verifyTokenURL := fmt.Sprintf("%s/verify-token", identityAPIEndpoint)
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			if includeIdentity(request) {
				ct := request.Context()
				token, err := getBearerToken(request)
				if err != nil {
					if err.Code == errs.NotFound {
						dataCollector.Logger.WarningWithContext(ct, "access token not found")
					} else {
						dataCollector.Logger.ErrorWithContext(ct, err)
					}
				} else {
					updatedCt, err := ctxWithUserID(dataCollector, httpClient, verifyTokenURL, ct, token)
					if err != nil {
						dataCollector.Logger.ErrorWithContext(ct, err)
					} else {
						request = request.WithContext(updatedCt)
					}
				}
			}

			handlerFunc(writer, request)
		}
	}
}

func ctxWithUserID(
	dataCollector telemetry.DataCollector,
	httpClient web.Client,
	verifyTokenURL string,
	ct context.Context,
	accessToken string) (context.Context, *errs.Error) {
	dataCollector.Logger.DebugWithContext(ct, "enter ctxWithUserID")
	defer dataCollector.Logger.DebugWithContext(ct, "exit ctxWithUserID")

	req, err := http.NewRequest(http.MethodPost, verifyTokenURL, bytes.NewReader([]byte(accessToken)))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	req.Header.Set("Content-Type", "text/plain")
	res, err := httpClient.Do(ct, req)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	internalErr := errs.GetFromHTTPErr(res)
	if internalErr != nil {
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	userID, err := strconv.ParseUint(string(buf), 10, 64)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.InvalidFormat,
			EmbedErr: err,
		}
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	return ctx.NewContextWithUserID(ct, userID), nil
}
