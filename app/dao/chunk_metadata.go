package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type ChunkMetadata interface {
	FindChunkMetadataID(ct context.Context, chunkID uint64) (entity.ChunkMetadata, error)
	CreateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) error
	UpdateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) error
}
