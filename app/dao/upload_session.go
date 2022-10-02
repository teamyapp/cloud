package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type UploadSession interface {
	FindUploadSessionByID(ct context.Context, uploadSessionID uint64) (entity.UploadSession, error)
	CreateUploadSession(ct context.Context, uploadSession entity.UploadSession) error
	UpdateUploadSession(ct context.Context, uploadSession entity.UploadSession) error
}
