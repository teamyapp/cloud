package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type UploadSession struct {
	db *dbtest.InMemoryDB
}

var _ dao.UploadSession = (*UploadSession)(nil)

func (u UploadSession) FindUploadSessionByID(ct context.Context, uploadSessionID uint64) (entity.UploadSession, *errs.Error) {
	table, err := u.db.GetTable(UploadSessionTableName)
	if err != nil {
		return entity.UploadSession{}, err
	}

	for _, rawRow := range table.Rows {
		uploadSession := rawRow.(entity.UploadSession)
		if uploadSession.ID == uploadSessionID {
			return uploadSession, nil
		}
	}

	return entity.UploadSession{}, errs.NewError(
		errs.NotFound,
		fmt.Sprintf("row not found: uploadSessionID=%v", uploadSessionID))
}

func (u UploadSession) CreateUploadSession(ct context.Context, uploadSession entity.UploadSession) *errs.Error {
	_, err := u.FindUploadSessionByID(ct, uploadSession.ID)
	if err == nil {
		return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: id=%v", uploadSession.ID))
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := u.db.GetTable(UploadSessionTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, uploadSession)
	return nil
}

func (u UploadSession) UpdateUploadSession(ct context.Context, uploadSession entity.UploadSession) *errs.Error {
	table, err := u.db.GetTable(UploadSessionTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		currUploadSession := rawRow.(entity.UploadSession)
		if currUploadSession.ID == uploadSession.ID {
			rows = append(rows, uploadSession)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.Rows = rows
		return nil
	}

	return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: id=%v", uploadSession.ID))
}

func NewUploadSession(db *dbtest.InMemoryDB) UploadSession {
	return UploadSession{
		db: db,
	}
}
