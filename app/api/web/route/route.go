package route

import (
	"net/http"
	"net/url"
	"path"
)

type Route struct {
	Path       string
	Method     string
	HandleFunc http.HandlerFunc
}

func WithChildPath(baseURL string, childPath string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, childPath)
	return u.String(), nil
}
