package service

import (
	"context"
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"sync"
	"time"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	tmio "github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type File struct {
	logger             telemetry.Logger
	mapClient          storage.MapClient
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
	now := time.Now().UTC()
	uploadSession.HashState = hashState
	uploadSession.UpdatedAt = &now
	internalErr = f.uploadSessionDao.UpdateUploadSession(ct, uploadSession)
	if internalErr != nil {
		return entity.UploadSession{}, internalErr
	}

	uploadSession.HashState = nil
	return uploadSession, nil
}

func (f File) AddChunk(ct context.Context, uploadSessionID uint64, chunkData io.Reader, contentLength int64) (entity.UploadSession, *errs.Error) {
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

	now := time.Now().UTC()
	chunkMetadata := entity.ChunkMetadata{
		ID:          chunkID,
		SizeInBytes: uint64(contentLength),
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

	readers := tmio.GenerateMultiReaders(chunkData, 2)
	hashReader, chunkReader := readers[0], readers[1]
	wg := sync.WaitGroup{}
	var wgErr *errs.Error
	once := sync.Once{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		internalErr = saveChunk(f.mapClient, chunkID, chunkReader)
		if internalErr != nil {
			once.Do(func() {
				wgErr = internalErr
			})
		}
	}()

	go func() {
		defer wg.Done()
		_, err = io.Copy(hashBuffer, hashReader)
		if err != nil {
			once.Do(func() {
				wgErr = errs.NewError(errs.IO, err.Error())
			})
		}
	}()

	wg.Wait()
	if wgErr != nil {
		return entity.UploadSession{}, wgErr
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

	chunksIterator := newChunksIterator(f.logger, f.mapClient, metadata.ChunkIDs)
	chunksBufferReader, chunksBufferWriter := io.Pipe()

	go func() {
		for {
			hasNext, err := chunksIterator.HasNext()
			if err != nil {
				f.logger.ErrorWithContext(ct, errs.NewError(errs.IO, "fail to invoke hasNext"))
				chunksBufferWriter.CloseWithError(err.ToError())
				return
			}

			if !hasNext {
				chunksBufferWriter.Close()
				return
			}

			data, err := chunksIterator.Next(ct)
			if err != nil {
				f.logger.ErrorWithContext(ct, errs.NewError(errs.IO, "failed to read chunk data"))
				chunksBufferWriter.CloseWithError(err.ToError())
				return
			}

			_, error := io.Copy(chunksBufferWriter, data)
			if error != nil {
				f.logger.ErrorWithContext(ct, errs.NewError(errs.IO, "failed to write chunk data"))
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

func NewFile(
	logger telemetry.Logger,
	mapClient storage.MapClient,
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
		mapClient:          mapClient,
		uploadSessionDao:   uploadSessionDao,
		fileMetadataDao:    fileMetadataDao,
		chunkMetadataDao:   chunkMetadataDao,
		uploadSessionIDGen: uploadSessionIDGen,
		chunkIDGen:         chunkIDGen,
		fileIDGen:          fileIDGen,
	}, nil
}
