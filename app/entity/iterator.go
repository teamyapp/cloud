package entity

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type Iterator[Item any] interface {
	HasNext() (bool, *errs.Error)
	Next(ct context.Context) (Item, *errs.Error)
}
