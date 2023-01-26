package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type SignInSession interface {
	FindSignInSessionByID(ct context.Context, sessionID uint64) (entity.SignInSession, error)
	CreateSignInSession(ct context.Context, session entity.SignInSession) error
	UpdateSignInSession(ct context.Context, session entity.SignInSession) error
	DeleteSignInSession(ct context.Context, sessionID uint64) error
}
