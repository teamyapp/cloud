package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type ServiceAccount struct {
	db *dbtest.InMemoryDB
}

var _ dao.ServiceAccount = (*ServiceAccount)(nil)

func (s ServiceAccount) FindAllServiceAccounts(ct context.Context, accountOwnerID uint64) ([]entity.ServiceAccount, *errs.Error) {
	table, err := s.db.GetTable(ServiceAccountTableName)
	if err != nil {
		return nil, err
	}

	serviceAccounts := make([]entity.ServiceAccount, 0)
	for _, rawRow := range table.Rows {
		serviceAccount := rawRow.(entity.ServiceAccount)
		if serviceAccount.OwnerUserID == accountOwnerID {
			serviceAccounts = append(serviceAccounts, serviceAccount)
		}
	}

	return serviceAccounts, nil
}

func (s ServiceAccount) FindServiceAccountByID(ct context.Context, serviceAccountID uint64) (entity.ServiceAccount, *errs.Error) {
	table, err := s.db.GetTable(ServiceAccountTableName)
	if err != nil {
		return entity.ServiceAccount{}, err
	}

	for _, rawRow := range table.Rows {
		serviceAccount := rawRow.(entity.ServiceAccount)
		if serviceAccount.ID == serviceAccountID {
			return serviceAccount, nil
		}
	}

	return entity.ServiceAccount{}, errs.NewError(
		errs.NotFound,
		fmt.Sprintf("row not found: serviceAccountID=%v", serviceAccountID))
}

func (s ServiceAccount) CreateServiceAccount(ct context.Context, serviceAccount entity.ServiceAccount) *errs.Error {
	_, err := s.FindServiceAccountByID(ct, serviceAccount.ID)
	if err == nil {
		return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: id=%v", serviceAccount.ID))
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := s.db.GetTable(ServiceAccountTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, serviceAccount)
	return nil
}

func (s ServiceAccount) UpdateServiceAccount(ct context.Context, serviceAccount entity.ServiceAccount) *errs.Error {
	table, err := s.db.GetTable(ServiceAccountTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		currServiceAccount := rawRow.(entity.ServiceAccount)
		if currServiceAccount.ID == serviceAccount.ID {
			rows = append(rows, serviceAccount)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.Rows = rows
		return nil
	}

	return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: id=%v", serviceAccount.ID))
}

func (s ServiceAccount) DeleteServiceAccount(ct context.Context, serviceAccountID uint64) *errs.Error {
	table, err := s.db.GetTable(ServiceAccountTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		serviceAccount := rawRow.(entity.ServiceAccount)
		if serviceAccount.ID != serviceAccountID {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewServiceAccount(db *dbtest.InMemoryDB) ServiceAccount {
	return ServiceAccount{
		db: db,
	}
}
