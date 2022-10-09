package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const gRPCAuthorizationKey = "Authorization"

func ServerHTTPWithIdentity(
	dataCollector obs.DataCollector,
	identityAPIEndpoint string,
) HTTPServerMiddleware {
	return withIdentity(dataCollector, identityAPIEndpoint, func(request *http.Request) (string, error) {
		value := request.Header.Get("Authorization")
		if len(value) == 0 {
			return "", nil
		}

		ct := request.Context()
		parts := strings.Split(value, " ")
		if len(parts) != 2 {
			err := errors.New("invalid Authorization header format")
			dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
				obs.CauseProp: err,
				obs.MessageProp: obs.Props{
					"format": value,
				},
			})
			return "", err
		}

		if parts[0] != "Bearer" {
			err := errors.New("invalid Authorization header format")
			dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
				obs.CauseProp: err,
				"format":      value,
			})
			return "", err
		}
		return parts[1], nil
	})
}

func ServerWebSocketWithIdentity(
	dataCollector obs.DataCollector,
	identityAPIEndpoint string,
) HTTPServerMiddleware {
	return withIdentity(dataCollector, identityAPIEndpoint, func(request *http.Request) (string, error) {
		return request.URL.Query().Get("accessToken"), nil
	})
}

func ServerGRPCWithIdentity(dataCollector obs.DataCollector, identityAPIEndpoint string) grpc.UnaryServerInterceptor {
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
				dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return nil, err
			}

			ct = updatedCt
		}

		return handler(ct, req)
	}
}

func ClientGRPCWithIdentity(getAccessToken func() string) grpc.UnaryClientInterceptor {
	return func(ct context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if getAccessToken != nil {
			ct = metadata.AppendToOutgoingContext(ct, gRPCAuthorizationKey, getAccessToken())
		}

		return invoker(ct, method, req, reply, cc, opts...)
	}
}

func withIdentity(
	dataCollector obs.DataCollector,
	identityAPIEndpoint string,
	getBearerToken func(request *http.Request) (string, error),
) HTTPServerMiddleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		verifyTokenURL := fmt.Sprintf("%s/verify-token", identityAPIEndpoint)
		return func(writer http.ResponseWriter, request *http.Request) {
			ct := request.Context()
			token, err := getBearerToken(request)
			if err != nil {
				dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			if len(token) > 0 {
				ct, err = ctxWithUserID(dataCollector, request.Context(), verifyTokenURL, token)
				if err != nil {
					dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
					writer.WriteHeader(http.StatusUnauthorized)
					return
				}

				request = request.WithContext(ct)
			}

			handlerFunc(writer, request)
		}
	}
}

func ctxWithUserID(dataCollector obs.DataCollector, ct context.Context, verifyTokenURL string, accessToken string) (context.Context, error) {
	res, err := http.Post(
		verifyTokenURL,
		"text/plain",
		bytes.NewReader([]byte(accessToken)))
	if err != nil {
		dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		err = errors.New("invalid access token")
		dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseUint(string(buf), 10, 64)
	if err != nil {
		dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return ctx.NewContextWithUserID(ct, userID), nil
}
