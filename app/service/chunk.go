package service

import (
	"context"
	"fmt"
	"io"
	"path"
	"strconv"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const chunkKeyPrefix = "chunks"

type ChunksIterator struct {
	logger         telemetry.Logger
	mapClient      storage.MapClient
	chunkIDs       []uint64
	nextChunkIndex int
}

var _ entity.Iterator[io.Reader] = (*ChunksIterator)(nil)

func (c *ChunksIterator) HasNext() (bool, *errs.Error) {
	return c.nextChunkIndex < len(c.chunkIDs), nil
}

func (c *ChunksIterator) Next(ct context.Context) (io.Reader, *errs.Error) {
	hasNext, err := c.HasNext()
	if err != nil {
		return nil, err
	}

	if !hasNext {
		return nil, errs.NewError(
			errs.InvalidOperation,
			fmt.Sprintf("no next chunk: nextChunkIndex=%v, numOfChunks=%v", c.nextChunkIndex, c.chunkIDs))
	}

	chunkIDPath := strconv.FormatUint(c.chunkIDs[c.nextChunkIndex], 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	data, err := c.mapClient.Get(fullPath)
	if err != nil {
		return nil, err
	}

	c.nextChunkIndex++
	return data, nil
}

func newChunksIterator(
	logger telemetry.Logger,
	mapClient storage.MapClient,
	chunkIDs []uint64,
) *ChunksIterator {
	return &ChunksIterator{
		logger:         logger,
		mapClient:      mapClient,
		chunkIDs:       chunkIDs,
		nextChunkIndex: 0,
	}
}

func saveChunk(mapClient storage.MapClient, chunkID uint64, data io.Reader) *errs.Error {
	chunkIDPath := strconv.FormatUint(chunkID, 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	return mapClient.Put(fullPath, data)
}
