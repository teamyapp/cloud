package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type ServiceAccount interface {
	FindAllServiceAccounts(ct context.Context, accountOwnerID uint64) ([]entity.ServiceAccount, error)
	FindServiceAccountByID(ct context.Context, serviceAccountID uint64) (entity.ServiceAccount, error)
	CreateServiceAccount(ct context.Context, serviceAccount entity.ServiceAccount) error
	UpdateServiceAccount(ct context.Context, serviceAccount entity.ServiceAccount) error
	DeleteServiceAccount(ct context.Context, serviceAccountID uint64) error
}
