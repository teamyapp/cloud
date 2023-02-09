package sqldb

import (
	"context"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

func parseIDs(ct context.Context, dataCollector telemetry.DataCollector, idsString string) ([]uint64, *errs.Error) {
	chunkIDs := make([]uint64, 0)
	if len(idsString) == 0 {
		return chunkIDs, nil
	}

	chunkIDStrings := strings.Split(idsString, ",")
	for _, chunkIDString := range chunkIDStrings {
		chunkID, err := strconv.ParseUint(chunkIDString, 10, 64)
		if err != nil {
			internalErr := &errs.Error{
				Code:     errs.InvalidArgument,
				EmbedErr: err,
			}
			dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return chunkIDs, internalErr
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
