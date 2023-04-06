package storage

import (
	"bytes"
	"io"
	"path"

	"github.com/minio/minio-go"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const appDataRoot = "appData"

type S3Bucket struct {
	logger     telemetry.Logger
	client     *minio.Client
	env        env.Environment
	bucketName string
}

var _ MapBackend = (*S3Bucket)(nil)

func (s S3Bucket) Get(key string) ([]byte, *errs.Error) {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	obj, err := s.client.GetObject(s.bucketName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	buf, err := io.ReadAll(obj)
	if err != nil {
		return nil, errs.NewError(errs.IO, err.Error())
	}

	return buf, nil
}

func (s S3Bucket) Put(key string, data []byte) *errs.Error {
	objSize := int64(len(data))
	fullPath := path.Join(appDataRoot, string(s.env), key)
	_, err := s.client.PutObject(s.bucketName, fullPath, bytes.NewReader(data), objSize, minio.PutObjectOptions{})
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s S3Bucket) Delete(key string) *errs.Error {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	err := s.client.RemoveObject(s.bucketName, fullPath)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewS3Bucket(
	logger telemetry.Logger,
	endpoint string,
	accessKeyID string,
	accessKey string,
	env env.Environment,
	bucketName string,
) (S3Bucket, error) {
	client, err := minio.New(endpoint, accessKeyID, accessKey, true)
	if err != nil {
		return S3Bucket{}, err
	}

	return S3Bucket{
		logger:     logger,
		client:     client,
		env:        env,
		bucketName: bucketName,
	}, nil
}
