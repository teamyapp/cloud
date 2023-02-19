package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

func WriteJSONToResponse(ct context.Context, dataCollector telemetry.DataCollector, writer http.ResponseWriter, body interface{}) *errs.Error {
	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	writer.Write(buf)
	return nil
}
