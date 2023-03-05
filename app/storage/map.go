package storage

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type MapBackend interface {
	Get(key string) ([]byte, *errs.Error)
	Put(key string, data []byte) *errs.Error
	Delete(key string) *errs.Error
}
