package storage

import (
	"context"
	"io"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
)

type Metadata struct {
	ContentType  string
	ETag         string
	LastModified time.Time
	Size         int64
	Name         string
}

type FileStream struct {
	Reader   io.Reader
	Metadata Metadata
}

type MapClient interface {
	Get(ct context.Context, key string) (io.Reader, *errs.Error)
	GetFileStreams(ct context.Context, key string) ([]FileStream, *errs.Error)
	Put(ct context.Context, key string, reader io.Reader) *errs.Error
	Delete(ct context.Context, key string) *errs.Error
	GetMetadata(ct context.Context, key string) (Metadata, *errs.Error)
}

type MapRequestHandlers interface {
	HandleGet(ct context.Context, key string) (io.Reader, *errs.Error)
	HandlePut(ct context.Context, key string, reader io.Reader) *errs.Error
	HandleDelete(ct context.Context, key string) *errs.Error
}
