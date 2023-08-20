package storage

import (
	"io"

	"github.com/teamyapp/cloud/libs/errs"
)

type MapClient interface {
	Get(key string) (io.Reader, *errs.Error)
	Put(key string, reader io.Reader) *errs.Error
	Delete(key string) *errs.Error
}

type MapRequestHandlers interface {
	HandleGet(key string) (io.Reader, *errs.Error)
	HandlePut(key string, reader io.Reader) *errs.Error
	HandleDelete(key string) *errs.Error
}
