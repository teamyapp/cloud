package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type ChunkMetadata struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.ChunkMetadata = (*ChunkMetadata)(nil)

func (c ChunkMetadata) FindChunkMetadataID(ct context.Context, chunkID uint64) (entity.ChunkMetadata, error) {
	chunkMetadata := entity.ChunkMetadata{}
	err := c.db.QueryRow(`
	SELECT
	    id,
	    size_in_bytes,
	    created_at
	FROM file_chunk_metadata
	WHERE id = $1;`,
		chunkID).
		Scan(
			&chunkMetadata.ID,
			&chunkMetadata.SizeInBytes,
			&chunkMetadata.CreatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.ChunkMetadata{}, dao.ErrNotFound(fmt.Sprintf(
			"chunk metadata not found: id=%v", chunkID))
	}

	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return chunkMetadata, err
}

func (c ChunkMetadata) CreateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) error {
	_, err := c.db.Exec(`
	INSERT INTO file_chunk_metadata
	(
	 	id,
	 	size_in_bytes,
	 	created_at
	)
	VALUES ($1, $2, $3);`,
		metadata.ID,
		metadata.SizeInBytes,
		metadata.CreatedAt,
	)

	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (c ChunkMetadata) UpdateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) error {
	_, err := c.db.Exec(`
	UPDATE file_chunk_metadata
	SET
	    id = $1,
	    size_in_bytes = $2,
	    created_at = $3
	WHERE id = $4;
	`,
		metadata.ID,
		metadata.SizeInBytes,
		metadata.CreatedAt,
		metadata.ID,
	)

	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewChunkMetadata(dataCollector telemetry.DataCollector, sqlDB *sql.DB) ChunkMetadata {
	return ChunkMetadata{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
