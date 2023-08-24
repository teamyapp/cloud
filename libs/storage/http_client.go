package storage

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/teamyapp/cloud/libs/errs"
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

	url := getFileURL(c.mapServerURL, fileID)
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
	url := getUploadFileUrl(c.mapServerURL, key)
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

func getFileURL(mapServerURL string, fileID uint64) string {
	fileIDParam := strconv.FormatUint(fileID, 10)
	return fmt.Sprintf("%s/files/%s", mapServerURL, fileIDParam)
}

func getUploadFileUrl(mapServerURL string, fileName string) string {
	return fmt.Sprintf("%s/files/upload?fileName=%s", mapServerURL, fileName)
}

func NewHTTPClient(mapServerURL string) *HTTPClient {
	return &HTTPClient{
		mapServerURL: mapServerURL,
	}
}
