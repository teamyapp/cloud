package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type FileMetadata struct {
	db *InMemoryDB
}

var _ dao.FileMetadata = (*FileMetadata)(nil)

func (f FileMetadata) FindMetadataByFileID(ct context.Context, fileID uint64) (entity.FileMetadata, *errs.Error) {
	table, err := f.db.GetTable(FileMetadataTableName)
	if err != nil {
		return entity.FileMetadata{}, err
	}

	for _, rawRow := range table.rows {
		fileMetadata := rawRow.(entity.FileMetadata)
		if fileMetadata.ID == fileID {
			return fileMetadata, nil
		}
	}

	return entity.FileMetadata{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: chunkID=%v", fileID),
	}
}

func (f FileMetadata) CreateFileMetadata(ct context.Context, metadata entity.FileMetadata) *errs.Error {
	_, err := f.FindMetadataByFileID(ct, metadata.ID)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: id=%v", metadata.ID),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := f.db.GetTable(FileMetadataTableName)
	if err != nil {
		return err
	}

	table.rows = append(table.rows, metadata)
	return nil
}

func (f FileMetadata) UpdateFileMetadata(ct context.Context, metadata entity.FileMetadata) *errs.Error {
	table, err := f.db.GetTable(FileMetadataTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		currFileMetadata := rawRow.(entity.FileMetadata)
		if currFileMetadata.ID == metadata.ID {
			rows = append(rows, metadata)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.rows = rows
		return nil
	}

	return &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: id=%v", metadata.ID),
	}
}

func NewFileMetadata(db *InMemoryDB) FileMetadata {
	return FileMetadata{
		db: db,
	}
}
