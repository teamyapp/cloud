package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type ServiceAccount interface {
	FindAllServiceAccounts(ct context.Context, accountOwnerID uint64) ([]entity.ServiceAccount, *errs.Error)
	FindServiceAccountByID(ct context.Context, serviceAccountID uint64) (entity.ServiceAccount, *errs.Error)
	CreateServiceAccount(ct context.Context, serviceAccount entity.ServiceAccount) *errs.Error
	UpdateServiceAccount(ct context.Context, serviceAccount entity.ServiceAccount) *errs.Error
	DeleteServiceAccount(ct context.Context, serviceAccountID uint64) *errs.Error
}
