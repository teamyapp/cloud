package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type UploadSession struct {
	db *InMemoryDB
}

var _ dao.UploadSession = (*UploadSession)(nil)

func (u UploadSession) FindUploadSessionByID(ct context.Context, uploadSessionID uint64) (entity.UploadSession, *errs.Error) {
	table, err := u.db.GetTable(UploadSessionTableName)
	if err != nil {
		return entity.UploadSession{}, err
	}

	for _, rawRow := range table.rows {
		uploadSession := rawRow.(entity.UploadSession)
		if uploadSession.ID == uploadSessionID {
			return uploadSession, nil
		}
	}

	return entity.UploadSession{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: uploadSessionID=%v", uploadSessionID),
	}
}

func (u UploadSession) CreateUploadSession(ct context.Context, uploadSession entity.UploadSession) *errs.Error {
	_, err := u.FindUploadSessionByID(ct, uploadSession.ID)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: uploadSession=%v", uploadSession),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := u.db.GetTable(UploadSessionTableName)
	if err != nil {
		return err
	}

	table.rows = append(table.rows, uploadSession)
	return nil
}

func (u UploadSession) UpdateUploadSession(ct context.Context, uploadSession entity.UploadSession) *errs.Error {
	table, err := u.db.GetTable(UploadSessionTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		currUploadSession := rawRow.(entity.UploadSession)
		if currUploadSession.ID == uploadSession.ID {
			rows = append(rows, uploadSession)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.rows = rows
		return nil
	}

	return &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: id=%v", uploadSession.ID),
	}
}

func NewUploadSession(db *InMemoryDB) UploadSession {
	return UploadSession{
		db: db,
	}
}
