package storage

import (
	"context"
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

var _ ObjectStore = (*HTTPClient)(nil)

func (*HTTPClient) GetDataStreams(ct context.Context, key string) ([]DataStream, *errs.Error) {
	panic("unimplemented")
}

func (*HTTPClient) GetMetadata(ct context.Context, key string) (Metadata, *errs.Error) {
	panic("unimplemented")
}

func (c *HTTPClient) Get(ct context.Context, key string) (io.Reader, *errs.Error) {
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

func (c *HTTPClient) Put(ct context.Context, key string, value io.Reader) *errs.Error {
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

func (c *HTTPClient) Delete(ct context.Context, key string) *errs.Error {
	panic("implement me")
}

func getFileURL(mapServerURL string, fileID uint64) string {
	fileIDParam := strconv.FormatUint(fileID, 10)
	return fmt.Sprintf("%s/files/%s", mapServerURL, fileIDParam)
}

func getUploadFileURL(mapServerURL string, fileName string) (string, *errs.Error) {
	uploadFileURL, err := url.Parse(fmt.Sprintf("%s/files/upload", mapServerURL))
	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	query := uploadFileURL.Query()
	query.Add("fileName", fileName)
	uploadFileURL.RawQuery = query.Encode()
	return uploadFileURL.String(), nil
}

func NewHTTPClient(mapServerURL string) *HTTPClient {
	return &HTTPClient{
		mapServerURL: mapServerURL,
	}
}
