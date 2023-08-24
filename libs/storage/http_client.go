package storage

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	url, err := getUploadFileURL(c.mapServerURL, key)
	if err != nil {
		return err
	}

	req, internalErr := http.NewRequest(http.MethodPost, url, value)
	if internalErr != nil {
		return errs.NewError(errs.Unknown, internalErr.Error())
	}

	_, internalErr = http.DefaultClient.Do(req)
	if internalErr != nil {
		return errs.NewError(errs.Unknown, internalErr.Error())
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

func getUploadFileURL(mapServerURL string, fileName string) (string, *errs.Error) {
	uploadFileUrl, err := url.Parse(fmt.Sprintf("%s/files/upload", mapServerURL))
	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	query := uploadFileUrl.Query()
	query.Add("fileName", fileName)
	uploadFileUrl.RawQuery = query.Encode()
	return uploadFileUrl.String(), nil
}

func NewHTTPClient(mapServerURL string) *HTTPClient {
	return &HTTPClient{
		mapServerURL: mapServerURL,
	}
}
