package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type ChunkMetadata struct {
	db *dbtest.InMemoryDB
}

var _ dao.ChunkMetadata = (*ChunkMetadata)(nil)

func (c ChunkMetadata) FindChunkMetadataID(ct context.Context, chunkID uint64) (entity.ChunkMetadata, *errs.Error) {
	table, err := c.db.GetTable(ChunkMetadataTableName)
	if err != nil {
		return entity.ChunkMetadata{}, err
	}

	for _, rawRow := range table.Rows {
		fileChunkMetadata := rawRow.(entity.ChunkMetadata)
		if fileChunkMetadata.ID == chunkID {
			return fileChunkMetadata, nil
		}
	}

	return entity.ChunkMetadata{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: chunkID=%v", chunkID),
	}
}

func (c ChunkMetadata) CreateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) *errs.Error {
	_, err := c.FindChunkMetadataID(ct, metadata.ID)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: id=%v", metadata.ID),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := c.db.GetTable(ChunkMetadataTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, metadata)
	return nil
}

func (c ChunkMetadata) UpdateChunkMetadata(ct context.Context, metadata entity.ChunkMetadata) *errs.Error {
	table, err := c.db.GetTable(ChunkMetadataTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		currFileChunkMetadata := rawRow.(entity.ChunkMetadata)
		if currFileChunkMetadata.ID == metadata.ID {
			rows = append(rows, metadata)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.Rows = rows
		return nil
	}

	return &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: id=%v", metadata.ID),
	}
}

func NewChunkMetadata(db *dbtest.InMemoryDB) ChunkMetadata {
	return ChunkMetadata{
		db: db,
	}
}
