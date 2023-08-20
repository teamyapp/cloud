package storage

import (
	"io"
	"net/http"
	"strconv"

	"github.com/teamyapp/cloud/libs/errs"
	tmio "github.com/teamyapp/cloud/libs/io"
)

type HTTPClient struct {
	mapServerURL string
}

var _ MapClient = (*HTTPClient)(nil)

func (c *HTTPClient) Get(key string) (io.Reader, *errs.Error) {
	fileID, err := strconv.ParseUint(key, 10, 64)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	url := tmio.GetFileURL(c.mapServerURL, fileID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	return res.Body, nil
}

func (c *HTTPClient) Put(key string, value io.Reader) *errs.Error {
	url := tmio.GetUploadFileURL(c.mapServerURL, key)
	req, err := http.NewRequest(http.MethodPost, url, value)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (c *HTTPClient) Delete(key string) *errs.Error {
	panic("implement me")
}

func NewHTTPClient(mapServerURL string) *HTTPClient {
	return &HTTPClient{
		mapServerURL: mapServerURL,
	}
}
