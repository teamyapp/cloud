package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type AllocatedRange interface {
	FindAllocatedRangeByKey(ct context.Context, key string) (entity.AllocatedRange, *errs.Error)
	CreateAllocatedRange(ct context.Context, allocatedRange entity.AllocatedRange) *errs.Error
	UpdateAllocatedRange(ct context.Context, allocatedRange entity.AllocatedRange) *errs.Error
}
