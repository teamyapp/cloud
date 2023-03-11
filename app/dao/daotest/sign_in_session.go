package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type SignInSession struct {
	db *dbtest.InMemoryDB
}

var _ dao.SignInSession = (*SignInSession)(nil)

func (s SignInSession) FindSignInSessionByID(ct context.Context, sessionID uint64) (entity.SignInSession, *errs.Error) {
	table, err := s.db.GetTable(SignInSessionTableName)
	if err != nil {
		return entity.SignInSession{}, err
	}

	for _, rawRow := range table.Rows {
		signInSession := rawRow.(entity.SignInSession)
		if signInSession.ID == sessionID {
			return signInSession, nil
		}
	}

	return entity.SignInSession{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: sessionID=%v", sessionID),
	}
}

func (s SignInSession) CreateSignInSession(ct context.Context, session entity.SignInSession) *errs.Error {
	_, err := s.FindSignInSessionByID(ct, session.ID)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: session=%v", session),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := s.db.GetTable(SignInSessionTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, session)
	return nil
}

func (s SignInSession) UpdateSignInSession(ct context.Context, session entity.SignInSession) *errs.Error {
	table, err := s.db.GetTable(SignInSessionTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		currSignInSession := rawRow.(entity.SignInSession)
		if currSignInSession.ID == session.ID {
			rows = append(rows, session)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.Rows = rows
		return nil
	}

	return &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: id=%v", session.ID),
	}
}

func (s SignInSession) DeleteSignInSession(ct context.Context, sessionID uint64) *errs.Error {
	table, err := s.db.GetTable(SignInSessionTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		signInSession := rawRow.(entity.SignInSession)
		if signInSession.ID != sessionID {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewSignInSession(db *dbtest.InMemoryDB) SignInSession {
	return SignInSession{
		db: db,
	}
}
