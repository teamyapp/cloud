package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type UploadSession interface {
	FindUploadSessionByID(ct context.Context, uploadSessionID uint64) (entity.UploadSession, *errs.Error)
	CreateUploadSession(ct context.Context, uploadSession entity.UploadSession) *errs.Error
	UpdateUploadSession(ct context.Context, uploadSession entity.UploadSession) *errs.Error
}
