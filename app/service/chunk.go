package service

import (
	"context"
	"fmt"
	"path"
	"strconv"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/storage"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const chunkKeyPrefix = "chunks"

type ChunksIterator struct {
	dataCollector  telemetry.DataCollector
	mapBackend     storage.MapBackend
	chunkIDs       []uint64
	nextChunkIndex int
}

var _ entity.Iterator[[]byte] = (*ChunksIterator)(nil)

func (c ChunksIterator) HasNext() (bool, *errs.Error) {
	return c.nextChunkIndex < len(c.chunkIDs), nil
}

func (c *ChunksIterator) Next(ct context.Context) ([]byte, *errs.Error) {
	hasNext, err := c.HasNext()
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	if !hasNext {
		err = &errs.Error{
			Code:    errs.InvalidOperation,
			Message: fmt.Sprintf("no next chunk: nextChunkIndex=%v, numOfChunks=%v", c.nextChunkIndex, c.chunkIDs),
		}
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	chunkIDPath := strconv.FormatUint(c.chunkIDs[c.nextChunkIndex], 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	data, err := c.mapBackend.Get(fullPath)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	c.nextChunkIndex++
	return data, nil
}

func newChunksIterator(
	dataCollector telemetry.DataCollector,
	mapBackend storage.MapBackend,
	chunkIDs []uint64,
) *ChunksIterator {
	return &ChunksIterator{
		dataCollector:  dataCollector,
		mapBackend:     mapBackend,
		chunkIDs:       chunkIDs,
		nextChunkIndex: 0,
	}
}

func saveChunk(mapBackend storage.MapBackend, chunkID uint64, data []byte) *errs.Error {
	chunkIDPath := strconv.FormatUint(chunkID, 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	return mapBackend.Put(fullPath, data)
}
