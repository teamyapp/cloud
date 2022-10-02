package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type FileMetadata interface {
	FindMetadataByFileID(ct context.Context, fileID uint64) (entity.FileMetadata, error)
	CreateFileMetadata(ct context.Context, metadata entity.FileMetadata) error
	UpdateFileMetadata(ct context.Context, metadata entity.FileMetadata) error
}
