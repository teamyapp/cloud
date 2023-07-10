package service

import (
	"context"
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/lang"
	"github.com/teamyapp/cloud/app/storage"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const chunksBufferSize = 10

type File struct {
	logger             telemetry.Logger
	mapBackend         storage.MapBackend
	uploadSessionDao   dao.UploadSession
	fileMetadataDao    dao.FileMetadata
	chunkMetadataDao   dao.ChunkMetadata
	uploadSessionIDGen *UniqueNumberGen
	chunkIDGen         *UniqueNumberGen
	fileIDGen          *UniqueNumberGen
}

func (f File) GetUploadSession(ct context.Context, uploadSessionID uint64) (entity.UploadSession, *errs.Error) {
	return f.uploadSessionDao.FindUploadSessionByID(ct, uploadSessionID)
}

func (f File) CreateUploadSession(ct context.Context) (uint64, *errs.Error) {
	chunkID, err := f.uploadSessionIDGen.GenerateUniqueNumber(ct)
	if err != nil {
		return 0, err
	}

	uploadSession := entity.UploadSession{
		ID:        chunkID,
		Status:    entity.CreatedUploadSessionStatus,
		CreatedAt: time.Now(),
	}

	err = f.uploadSessionDao.CreateUploadSession(ct, uploadSession)
	if err != nil {
		return 0, err
	}

	return chunkID, nil
}

func (f File) InitUploadSession(
	ct context.Context,
	uploadSessionID uint64,
	fileName string,
	mimeType string,
	expectedContentHash string,
	totalSizeInBytes uint64,
	totalNumOfChunks int,
) (entity.UploadSession, *errs.Error) {
	uploadSession, internalErr := f.uploadSessionDao.FindUploadSessionByID(ct, uploadSessionID)
	if internalErr != nil {
		return entity.UploadSession{}, internalErr
	}

	switch uploadSession.Status {
	case entity.CompletedUploadSessionStatus:
		return entity.UploadSession{}, errs.NewError(errs.InvalidOperation, "upload session is already completed")
	case entity.InitializedUploadSessionStatus, entity.UploadingChunksUploadSessionStatus:
		return entity.UploadSession{}, errs.NewError(errs.InvalidOperation, "upload session is already initialized")
	}

	hashBuffer := sha256.New()
	hashState, err := hashBuffer.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		return entity.UploadSession{}, errs.NewError(errs.Deserialization, err.Error())
	}

	uploadSession.FileName = fileName
	uploadSession.MIMEType = mimeType
	uploadSession.ExpectedContentHash = expectedContentHash
	uploadSession.TotalSizeInBytes = totalSizeInBytes
	uploadSession.TotalNumOfChunks = totalNumOfChunks
	uploadSession.Status = entity.InitializedUploadSessionStatus
	now := time.Now()
	uploadSession.HashState = hashState
	uploadSession.UpdatedAt = &now
	internalErr = f.uploadSessionDao.UpdateUploadSession(ct, uploadSession)
	if internalErr != nil {
		return entity.UploadSession{}, internalErr
	}

	uploadSession.HashState = nil
	return uploadSession, nil
}

func (f File) AddChunk(ct context.Context, uploadSessionID uint64, chunkData []byte) (entity.UploadSession, *errs.Error) {
	uploadSession, internalErr := f.uploadSessionDao.FindUploadSessionByID(ct, uploadSessionID)
	if internalErr != nil {
		return entity.UploadSession{}, internalErr
	}

	switch uploadSession.Status {
	case entity.CompletedUploadSessionStatus:
		return entity.UploadSession{}, errs.NewError(errs.InvalidOperation, "upload session is already completed")
	case entity.CreatedUploadSessionStatus:
		return entity.UploadSession{}, errs.NewError(errs.InvalidOperation, "upload session is not initialized")
	}

	chunkID, internalErr := f.chunkIDGen.GenerateUniqueNumber(ct)
	if internalErr != nil {
		return entity.UploadSession{}, internalErr
	}

	internalErr = saveChunk(f.mapBackend, chunkID, chunkData)
	if internalErr != nil {
		return entity.UploadSession{}, internalErr
	}

	now := time.Now()
	chunkMetadata := entity.ChunkMetadata{
		ID:          chunkID,
		SizeInBytes: uint64(len(chunkData)),
		CreatedAt:   now,
	}
	internalErr = f.chunkMetadataDao.CreateChunkMetadata(ct, chunkMetadata)
	if internalErr != nil {
		return entity.UploadSession{}, internalErr
	}

	hashBuffer := sha256.New()
	err := hashBuffer.(encoding.BinaryUnmarshaler).UnmarshalBinary(uploadSession.HashState)
	if err != nil {
		return entity.UploadSession{}, errs.NewError(errs.Deserialization, err.Error())
	}

	_, err = hashBuffer.Write(chunkData)
	if err != nil {
		return entity.UploadSession{}, errs.NewError(errs.IO, err.Error())
	}

	hashState, err := hashBuffer.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		return entity.UploadSession{}, errs.NewError(errs.Serialization, err.Error())
	}

	uploadSession.HashState = hashState
	uploadSession.ChunkIDs = append(uploadSession.ChunkIDs, chunkID)
	uploadSession.NextChunkIndexToUpload++
	uploadSession.UploadedSizeInBytes += chunkMetadata.SizeInBytes
	uploadSession.UpdatedAt = &now
	if uploadSession.NextChunkIndexToUpload < uploadSession.TotalNumOfChunks {
		uploadSession.Status = entity.UploadingChunksUploadSessionStatus
	} else {
		uploadSession, internalErr = f.FinishFileUpload(ct, uploadSession, hashBuffer)
		if internalErr != nil {
			return entity.UploadSession{}, internalErr
		}
	}

	internalErr = f.uploadSessionDao.UpdateUploadSession(ct, uploadSession)
	if internalErr != nil {
		return uploadSession, internalErr
	}

	uploadSession.HashState = nil
	return uploadSession, nil
}

func (f File) FinishFileUpload(ct context.Context, uploadSession entity.UploadSession, hashBuffer hash.Hash) (entity.UploadSession, *errs.Error) {
	uploadSession.Status = entity.CompletedUploadSessionStatus
	actualHash := hashBuffer.Sum(nil)
	actualHashString := hex.EncodeToString(actualHash)
	if actualHashString != uploadSession.ExpectedContentHash {
		return entity.UploadSession{}, errs.NewError(
			errs.InvalidOperation,
			fmt.Sprintf("sha256 hash not match: actualHash=%v, expectedHash=%v",
				actualHashString,
				uploadSession.ExpectedContentHash))
	}

	uploadSession.ActualContentHash = actualHashString
	fileID, err := f.fileIDGen.GenerateUniqueNumber(ct)
	if err != nil {
		return entity.UploadSession{}, err
	}

	fileMetadata := entity.FileMetadata{
		ID:          fileID,
		Name:        uploadSession.FileName,
		SizeInBytes: uploadSession.TotalSizeInBytes,
		MIMEType:    uploadSession.MIMEType,
		ChunkIDs:    uploadSession.ChunkIDs,
		CreatedAt:   uploadSession.CreatedAt,
	}
	uploadSession.FileID = fileID
	return uploadSession, f.fileMetadataDao.CreateFileMetadata(ct, fileMetadata)
}

func (f File) GetFileMetadata(ct context.Context, fileID uint64) (entity.FileMetadata, *errs.Error) {
	return f.fileMetadataDao.FindMetadataByFileID(ct, fileID)
}

func (f File) GetFile(ct context.Context, fileID uint64) (entity.File, *errs.Error) {
	metadata, err := f.GetFileMetadata(ct, fileID)
	if err != nil {
		return entity.File{}, err
	}

	chunksIterator := newChunksIterator(f.logger, f.mapBackend, metadata.ChunkIDs)
	chunksBuffer := make(chan lang.Result[[]byte], chunksBufferSize)
	go func() {
		defer close(chunksBuffer)
		for {
			hasNext, err := chunksIterator.HasNext()
			if err != nil {
				chunksBuffer <- lang.Result[[]byte]{
					Error: err,
				}
				return
			}

			if !hasNext {
				return
			}

			data, err := chunksIterator.Next(ct)
			if err != nil {
				chunksBuffer <- lang.Result[[]byte]{
					Error: err,
				}
				return
			}

			chunksBuffer <- lang.Result[[]byte]{
				Value: data,
				Error: nil,
			}

			// Required to prevent Wire gen from failure
			continue
		}
	}()
	return entity.File{
		Metadata:     metadata,
		ChunksBuffer: chunksBuffer,
	}, nil
}

func NewFile(
	logger telemetry.Logger,
	mapBackend storage.MapBackend,
	uniqueNumberRegistry *UniqueNumberGenRegistry,
	uploadSessionDao dao.UploadSession,
	fileMetadataDao dao.FileMetadata,
	chunkMetadataDao dao.ChunkMetadata,
) (File, error) {
	uploadSessionIDGen, err := uniqueNumberRegistry.GetUniqueNumberGen("uploadSessionID")
	if err != nil {
		return File{}, err.ToError()
	}

	chunkIDGen, err := uniqueNumberRegistry.GetUniqueNumberGen("chunkID")
	if err != nil {
		return File{}, err.ToError()
	}

	fileIDGen, err := uniqueNumberRegistry.GetUniqueNumberGen("fileID")
	if err != nil {
		return File{}, err.ToError()
	}

	return File{
		logger:             logger,
		mapBackend:         mapBackend,
		uploadSessionDao:   uploadSessionDao,
		fileMetadataDao:    fileMetadataDao,
		chunkMetadataDao:   chunkMetadataDao,
		uploadSessionIDGen: uploadSessionIDGen,
		chunkIDGen:         chunkIDGen,
		fileIDGen:          fileIDGen,
	}, nil
}
