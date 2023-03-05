package webtest

import (
	"net/http"

	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/web"
)

func InsecureHTTPClient(network network.Network) web.HTTPClient {
	rawHTTPClient := web.NewHTTPClient(network)
	httpClient := web.HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		return rawHTTPClient.Do(req)
	})
	return httpClient
}
