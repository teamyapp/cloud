package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type SignInSession interface {
	FindSignInSessionByID(ct context.Context, sessionID uint64) (entity.SignInSession, *errs.Error)
	CreateSignInSession(ct context.Context, session entity.SignInSession) *errs.Error
	UpdateSignInSession(ct context.Context, session entity.SignInSession) *errs.Error
	DeleteSignInSession(ct context.Context, sessionID uint64) *errs.Error
}
