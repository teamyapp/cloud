package web

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/teamyapp/cloud/libs/network"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HTTPClientFunc func(req *http.Request) (*http.Response, error)

var _ HTTPClient = (*HTTPClientFunc)(nil)

func (h HTTPClientFunc) Do(req *http.Request) (*http.Response, error) {
	return (h)(req)
}

func NewHTTPClient(nw network.Network) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return nw.Dial(network, addr)
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
