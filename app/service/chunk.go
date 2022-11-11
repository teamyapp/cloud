package service

import (
	"context"
	"errors"
	"path"
	"strconv"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/storage"
	"github.com/teamyapp/cloud/libs/obs"
)

const chunkKeyPrefix = "chunks"

type ChunksIterator struct {
	dataCollector  obs.DataCollector
	mapBackend     storage.MapBackend
	chunkIDs       []uint64
	nextChunkIndex int
}

var _ entity.Iterator[[]byte] = (*ChunksIterator)(nil)

func (c ChunksIterator) HasNext() (bool, error) {
	return c.nextChunkIndex < len(c.chunkIDs), nil
}

func (c *ChunksIterator) Next(ct context.Context) ([]byte, error) {
	hasNext, err := c.HasNext()
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if !hasNext {
		err = errors.New("no next chunk")
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp:    err,
			"NextChunkIndex": c.nextChunkIndex,
			"NumOfChunks":    c.chunkIDs,
		})
		return nil, err
	}

	chunkIDPath := strconv.FormatUint(c.chunkIDs[c.nextChunkIndex], 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	data, err := c.mapBackend.Get(fullPath)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	c.nextChunkIndex++
	return data, nil
}

func newChunksIterator(
	dataCollector obs.DataCollector,
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

func saveChunk(mapBackend storage.MapBackend, chunkID uint64, data []byte) error {
	chunkIDPath := strconv.FormatUint(chunkID, 10)
	fullPath := path.Join(chunkKeyPrefix, chunkIDPath)
	return mapBackend.Put(fullPath, data)
}
