package sqldb

import (
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/obs"
)

func parseIDs(dataCollector obs.DataCollector, idsString string) ([]uint64, error) {
	chunkIDs := make([]uint64, 0)
	if len(idsString) == 0 {
		return chunkIDs, nil
	}

	chunkIDStrings := strings.Split(idsString, ",")
	for _, chunkIDString := range chunkIDStrings {
		chunkID, err := strconv.ParseUint(chunkIDString, 10, 64)
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return chunkIDs, err
		}

		chunkIDs = append(chunkIDs, chunkID)
	}

	return chunkIDs, nil
}

func formatIDs(ids []uint64) string {
	chunkIDStrings := collect.Map(ids, func(id uint64, index int) string {
		return strconv.FormatUint(id, 10)
	})
	return strings.Join(chunkIDStrings, ",")
}
