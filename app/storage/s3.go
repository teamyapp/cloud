package storage

import (
	"bytes"
	"io"
	"path"

	"github.com/minio/minio-go"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/obs"
)

const appDataRoot = "appData"

type S3Bucket struct {
	dataCollector obs.DataCollector
	client        *minio.Client
	env           config.Environment
	bucketName    string
}

var _ MapBackend = (*S3Bucket)(nil)

func (s S3Bucket) Get(key string) ([]byte, error) {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	obj, err := s.client.GetObject(s.bucketName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return io.ReadAll(obj)
}

func (s S3Bucket) Put(key string, data []byte) error {
	objSize := int64(len(data))
	fullPath := path.Join(appDataRoot, string(s.env), key)
	_, err := s.client.PutObject(s.bucketName, fullPath, bytes.NewReader(data), objSize, minio.PutObjectOptions{})
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (s S3Bucket) Delete(key string) error {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	err := s.client.RemoveObject(s.bucketName, fullPath)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewS3Bucket(
	dataCollector obs.DataCollector,
	endpoint string,
	accessKeyID string,
	accessKey string,
	env config.Environment,
	bucketName string,
) (S3Bucket, error) {
	client, err := minio.New(endpoint, accessKeyID, accessKey, true)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return S3Bucket{}, err
	}

	return S3Bucket{
		dataCollector: dataCollector,
		client:        client,
		env:           env,
		bucketName:    bucketName,
	}, nil
}
