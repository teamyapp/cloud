package storage

import (
	"context"
	"io"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
)

type ObjectMetadata struct {
	ContentType  string
	ETag         string
	LastModified time.Time
	Size         int64
	Name         string
}

type ObjectDataStream struct {
	Reader   io.Reader
	Metadata ObjectMetadata
}

type ObjectStore interface {
	Get(ct context.Context, key string) (io.Reader, *errs.Error)
	GetDataStreams(ct context.Context, key string) ([]ObjectDataStream, *errs.Error)
	Put(ct context.Context, key string, reader io.Reader) *errs.Error
	Delete(ct context.Context, key string) *errs.Error
	GetMetadata(ct context.Context, key string) (ObjectMetadata, *errs.Error)
}

type ObjectStoreRequestHandlers interface {
	HandleGet(ct context.Context, key string) (io.Reader, *errs.Error)
	HandlePut(ct context.Context, key string, reader io.Reader) *errs.Error
	HandleDelete(ct context.Context, key string) *errs.Error
}
