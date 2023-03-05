package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type SignInSession struct {
	db *InMemoryDB
}

var _ dao.SignInSession = (*SignInSession)(nil)

func (s SignInSession) FindSignInSessionByID(ct context.Context, sessionID uint64) (entity.SignInSession, *errs.Error) {
	table, err := s.db.GetTable(SignInSessionTableName)
	if err != nil {
		return entity.SignInSession{}, err
	}

	for _, rawRow := range table.rows {
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

	table.rows = append(table.rows, session)
	return nil
}

func (s SignInSession) UpdateSignInSession(ct context.Context, session entity.SignInSession) *errs.Error {
	table, err := s.db.GetTable(SignInSessionTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		currSignInSession := rawRow.(entity.SignInSession)
		if currSignInSession.ID == session.ID {
			rows = append(rows, session)
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
		Message: fmt.Sprintf("row not found: id=%v", session.ID),
	}
}

func (s SignInSession) DeleteSignInSession(ct context.Context, sessionID uint64) *errs.Error {
	table, err := s.db.GetTable(SignInSessionTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		signInSession := rawRow.(entity.SignInSession)
		if signInSession.ID != sessionID {
			rows = append(rows, rawRow)
		}
	}

	table.rows = rows
	return nil
}

func NewSignInSession(db *InMemoryDB) SignInSession {
	return SignInSession{
		db: db,
	}
}
