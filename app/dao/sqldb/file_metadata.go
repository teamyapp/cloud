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

type FileMetadata struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.FileMetadata = (*FileMetadata)(nil)

func (f FileMetadata) FindMetadataByFileID(ct context.Context, fileID uint64) (entity.FileMetadata, error) {
	fileMetadata := entity.FileMetadata{}
	var chunkIDsString string
	err := f.db.QueryRow(`
	SELECT
	    id,
	    name,
	    size_in_bytes,
	    mime_type,
	    chunk_ids,
	    created_at,
	    last_modified_at
	FROM file_metadata
	WHERE id = $1;`,
		fileID).
		Scan(
			&fileMetadata.ID,
			&fileMetadata.Name,
			&fileMetadata.SizeInBytes,
			&fileMetadata.MIMEType,
			&chunkIDsString,
			&fileMetadata.CreatedAt,
			&fileMetadata.LastModifiedAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.FileMetadata{}, dao.ErrNotFound(fmt.Sprintf(
			"file metadata not found: id=%v", fileID))
	}

	if err != nil {
		f.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.FileMetadata{}, err
	}

	chunkIDs, err := parseIDs(ct, f.dataCollector, chunkIDsString)
	if err != nil {
		f.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.FileMetadata{}, err
	}

	fileMetadata.ChunkIDs = chunkIDs
	return fileMetadata, nil
}

func (f FileMetadata) CreateFileMetadata(ct context.Context, metadata entity.FileMetadata) error {
	_, err := f.db.Exec(`
	INSERT INTO file_metadata
	(
	 	id,
	 	name,
	 	size_in_bytes,
	 	mime_type,
	 	chunk_ids,
	 	created_at,
	 	last_modified_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		metadata.ID,
		metadata.Name,
		metadata.SizeInBytes,
		metadata.MIMEType,
		formatIDs(metadata.ChunkIDs),
		metadata.CreatedAt,
		metadata.LastModifiedAt,
	)
	if err != nil {
		f.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (f FileMetadata) UpdateFileMetadata(ct context.Context, metadata entity.FileMetadata) error {
	_, err := f.db.Exec(`
	UPDATE file_metadata
	SET
	    id = $1,
	    name = $2,
	    size_in_bytes = $3,
	    mime_type = $4,
	    chunk_ids = $5,
	    created_at = $6,
	    last_modified_at = $7
	WHERE id = $8;
	`,
		metadata.ID,
		metadata.Name,
		metadata.SizeInBytes,
		metadata.MIMEType,
		formatIDs(metadata.ChunkIDs),
		metadata.CreatedAt,
		metadata.LastModifiedAt,
	)

	if err != nil {
		f.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewFileMetadata(dataCollector telemetry.DataCollector, sqlDB *sql.DB) FileMetadata {
	return FileMetadata{dataCollector: dataCollector, db: sqlDB}
}
