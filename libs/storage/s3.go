package storage

import (
	"context"
	"io"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const appDataRoot = ""

type S3Bucket struct {
	logger     telemetry.Logger
	client     *minio.Client
	env        env.Environment
	bucketName string
}

var _ ObjectStore = (*S3Bucket)(nil)
var _ MapRequestHandlers = (*S3Bucket)(nil)

func (s *S3Bucket) GetMetadata(ct context.Context, key string) (Metadata, *errs.Error) {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	obj, err := s.client.StatObject(ct, s.bucketName, fullPath, minio.StatObjectOptions{})
	if err != nil {
		return Metadata{}, errs.NewError(errs.Unknown, err.Error())
	}

	return Metadata{
		ContentType:  obj.ContentType,
		ETag:         obj.ETag,
		LastModified: obj.LastModified,
		Size:         obj.Size,
		Name:         obj.Key,
	}, nil
}

func (s *S3Bucket) GetDataStreams(ct context.Context, key string) ([]DataStream, *errs.Error) {
	fileStreams := []DataStream{}
	// We need the slash at the end to make sure we only get the files in the directory
	prefix := path.Join(appDataRoot, string(s.env), key) + "/"

	for obj := range s.client.ListObjects(ct, s.bucketName, minio.ListObjectsOptions{
		Prefix:       prefix,
		WithMetadata: true,
		Recursive:    true,
	}) {
		if obj.Err != nil {
			return nil, errs.NewError(errs.Unknown, obj.Err.Error())
		}

		objReader, err := s.client.GetObject(ct, s.bucketName, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		fileStreams = append(fileStreams, DataStream{
			Reader: objReader,
			Metadata: Metadata{
				ContentType:  obj.ContentType,
				ETag:         obj.ETag,
				LastModified: obj.LastModified,
				Size:         obj.Size,
				Name:         obj.Key,
			},
		})
	}

	return fileStreams, nil
}

func (s *S3Bucket) Get(ct context.Context, key string) (io.Reader, *errs.Error) {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	obj, err := s.client.GetObject(ct, s.bucketName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	return obj, nil
}

func (s *S3Bucket) Put(ct context.Context, key string, data io.Reader) *errs.Error {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	_, err := s.client.PutObject(ct, s.bucketName, fullPath, data, -1, minio.PutObjectOptions{})
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *S3Bucket) Delete(ct context.Context, key string) *errs.Error {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	err := s.client.RemoveObject(ct, s.bucketName, fullPath, minio.RemoveObjectOptions{})
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *S3Bucket) HandleGet(ct context.Context, key string) (io.Reader, *errs.Error) {
	return s.Get(ct, key)
}

func (s *S3Bucket) HandlePut(ct context.Context, key string, data io.Reader) *errs.Error {
	return s.Put(ct, key, data)
}

func (s *S3Bucket) HandleDelete(ct context.Context, key string) *errs.Error {
	return s.Delete(ct, key)
}

func NewS3Bucket(
	logger telemetry.Logger,
	endpoint string,
	accessKeyID string,
	accessKey string,
	env env.Environment,
	bucketName string,
) (*S3Bucket, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, accessKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}

	return &S3Bucket{
		logger:     logger,
		client:     client,
		env:        env,
		bucketName: bucketName,
	}, nil
}
