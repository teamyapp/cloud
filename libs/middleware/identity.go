package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/app/ctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func withIdentity(
	identityAPIEndpoint string,
	handlerFunc http.HandlerFunc,
	getBearerToken func(request *http.Request) (string, error),
) http.HandlerFunc {
	verifyTokenURL := fmt.Sprintf("%s/verify-token", identityAPIEndpoint)
	return func(writer http.ResponseWriter, request *http.Request) {
		token, err := getBearerToken(request)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		if len(token) > 0 {
			ct, err := ctxWithUserID(request.Context(), verifyTokenURL, token)
			if err != nil {
				log.Println(err)
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			request = request.WithContext(ct)
		}

		handlerFunc(writer, request)
	}
}

func ServerWithWebIdentity(
	identityAPIEndpoint string,
	handlerFunc http.HandlerFunc,
) http.HandlerFunc {
	return withIdentity(identityAPIEndpoint, handlerFunc, func(request *http.Request) (string, error) {
		value := request.Header.Get("Authorization")
		if len(value) == 0 {
			return "", nil
		}

		parts := strings.Split(value, " ")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid Authorization header format: %s\n", value)
			log.Println(err)
			return "", err
		}

		if parts[0] != "Bearer" {
			err := fmt.Errorf("invalid Authorization header format: %v\n", parts)
			log.Println(err)
			return "", err
		}
		return parts[1], nil
	})
}

func ServerWithWebSocketIdentity(
	identityAPIEndpoint string,
	handlerFunc http.HandlerFunc,
) http.HandlerFunc {
	return withIdentity(identityAPIEndpoint, handlerFunc, func(request *http.Request) (string, error) {
		return request.URL.Query().Get("accessToken"), nil
	})
}

func ServerWithGRPCIdentity(verifyTokenURL string) grpc.UnaryServerInterceptor {
	return func(ct context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ct)
		if !ok {
			return handler(ct, req)
		}

		values := md.Get("Authorization")
		if len(values) > 0 {
			accessToken := values[0]
			ct, err = ctxWithUserID(ct, verifyTokenURL, accessToken)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}

		return handler(ct, req)
	}
}

func ctxWithUserID(ct context.Context, verifyTokenURL string, accessToken string) (context.Context, error) {
	res, err := http.Post(
		verifyTokenURL,
		"text/plain",
		bytes.NewReader([]byte(accessToken)))
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("invalid access token")
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseUint(string(buf), 10, 64)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ctx.NewContextWithUserID(ct, userID), err
}
