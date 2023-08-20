package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type FileMetadata interface {
	FindMetadataByFileID(ct context.Context, fileID uint64) (entity.FileMetadata, *errs.Error)
	CreateFileMetadata(ct context.Context, metadata entity.FileMetadata) *errs.Error
	UpdateFileMetadata(ct context.Context, metadata entity.FileMetadata) *errs.Error
}
