package web

import (
	"context"
	"net/http"
)

type HTTPClient func(ct context.Context, req *http.Request) (*http.Response, error)

func (h *HTTPClient) Do(ct context.Context, req *http.Request) (*http.Response, error) {
	return (*h)(ct, req)
}
