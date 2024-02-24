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

type DataStream struct {
	Reader   io.Reader
	Metadata Metadata
}

type ObjectStore interface {
	Get(ct context.Context, key string) (io.Reader, *errs.Error)
	GetDataStreams(ct context.Context, key string) ([]DataStream, *errs.Error)
	Put(ct context.Context, key string, reader io.Reader) *errs.Error
	Delete(ct context.Context, key string) *errs.Error
	GetMetadata(ct context.Context, key string) (Metadata, *errs.Error)
}

type MapRequestHandlers interface {
	HandleGet(ct context.Context, key string) (io.Reader, *errs.Error)
	HandlePut(ct context.Context, key string, reader io.Reader) *errs.Error
	HandleDelete(ct context.Context, key string) *errs.Error
}
