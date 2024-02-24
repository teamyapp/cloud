package service

import (
	"context"
	"io"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Stream struct {
	logger          telemetry.Logger
	objectStore     storage.ObjectStore
	fileMetadataDao dao.FileMetadata
}

func (s Stream) GetFileMetadata(ct context.Context, fileID uint64) (entity.FileMetadata, *errs.Error) {
	return s.fileMetadataDao.FindMetadataByFileID(ct, fileID)
}

func (s Stream) AddFile(ct context.Context, fileName string, fileData io.Reader) *errs.Error {
	return s.objectStore.Put(ct, fileName, fileData)
}

func (s Stream) GetFile(ct context.Context, fileID uint64) (entity.File, *errs.Error) {
	metadata, err := s.GetFileMetadata(ct, fileID)
	if err != nil {
		return entity.File{}, err
	}

	chunksIterator := newChunksIterator(s.logger, s.objectStore, metadata.ChunkIDs)
	chunksBufferReader, chunksBufferWriter := io.Pipe()

	go func() {
		for {
			hasNext, err := chunksIterator.HasNext()
			if err != nil {
				s.logger.ErrorWithContext(ct, errs.NewError(errs.IO, "failed to read has next"))
				chunksBufferWriter.CloseWithError(err.ToError())
				return
			}

			if !hasNext {
				chunksBufferWriter.Close()
				return
			}

			data, err := chunksIterator.Next(ct)
			if err != nil {
				s.logger.ErrorWithContext(ct, errs.NewError(errs.IO, "failed to read chunk data"))
				chunksBufferWriter.CloseWithError(err.ToError())
				return
			}

			_, error := io.Copy(chunksBufferWriter, data)
			if error != nil {
				s.logger.ErrorWithContext(ct, errs.NewError(errs.IO, "failed to write chunk data"))
				chunksBufferWriter.CloseWithError(error)
				return
			}
		}

	}()
	return entity.File{
		Metadata:     metadata,
		ChunksBuffer: chunksBufferReader,
	}, nil
}

func NewStream(logger telemetry.Logger, objectStore storage.ObjectStore, fileMetadataDao dao.FileMetadata) Stream {
	return Stream{
		logger:          logger,
		objectStore:     objectStore,
		fileMetadataDao: fileMetadataDao,
	}
}
