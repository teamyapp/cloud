package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type ChunkMetadata interface {
	FindChunkMetadataID(ct context.Context, chunkID uint64) (entity.ChunkMetadata, *errs.Error)
	CreateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) *errs.Error
	UpdateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) *errs.Error
}
