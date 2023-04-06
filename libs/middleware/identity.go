package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/identity"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const gRPCAuthorizationKey = "Authorization"

type IncludeIdentityWebFunc func(request *http.Request) bool
type IncludeIdentityGRPCFunc func(info *grpc.UnaryServerInfo) bool

func ServerHTTPWithIdentity(
	logger telemetry.Logger,
	httpClient web.HTTPClient,
	identityAPIEndpoint string,
	includeIdentity IncludeIdentityWebFunc,
) Middleware[http.HandlerFunc] {
	return withIdentity(
		logger,
		httpClient,
		identityAPIEndpoint,
		identity.GetBearerToken,
		includeIdentity)
}

func ServerWebSocketWithIdentity(
	logger telemetry.Logger,
	httpClient web.HTTPClient,
	identityAPIEndpoint string,
	includeIdentity IncludeIdentityWebFunc,
) Middleware[http.HandlerFunc] {
	return withIdentity(logger, httpClient, identityAPIEndpoint, func(request *http.Request) (string, *errs.Error) {
		token := request.URL.Query().Get("accessToken")
		if len(token) == 0 {
			return "", errs.NewError(errs.NotFound, "access token not found")
		}

		return token, nil
	}, includeIdentity)
}

func ServerGRPCWithIdentity(
	logger telemetry.Logger,
	httpClient web.HTTPClient,
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
				updatedCt, err := ctxWithUserID(logger, httpClient, verifyTokenURL, ct, accessToken)
				if err != nil {
					logger.WarningWithContext(ct, err.String())
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
	logger telemetry.Logger,
	httpClient web.HTTPClient,
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
						logger.WarningWithContext(ct, "access token not found")
					} else {
						logger.ErrorWithContext(ct, err)
					}
				} else {
					updatedCt, err := ctxWithUserID(logger, httpClient, verifyTokenURL, ct, token)
					if err != nil {
						logger.ErrorWithContext(ct, err)
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
	logger telemetry.Logger,
	httpClient web.HTTPClient,
	verifyTokenURL string,
	ct context.Context,
	accessToken string,
) (context.Context, *errs.Error) {
	logger.DebugWithContext(ct, "enter ctxWithUserID")
	defer logger.DebugWithContext(ct, "exit ctxWithUserID")

	req, err := http.NewRequest(http.MethodPost, verifyTokenURL, bytes.NewReader([]byte(accessToken)))
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(ct)
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	internalErr := errs.GetFromHTTPErr(res)
	if internalErr != nil {
		logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errs.NewError(errs.IO, err.Error())
	}

	userID, err := strconv.ParseUint(string(buf), 10, 64)
	if err != nil {
		return nil, errs.NewError(errs.InvalidFormat, err.Error())
	}

	return ctx.NewContextWithUserID(ct, userID), nil
}
