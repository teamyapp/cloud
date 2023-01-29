package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/teamyapp/cloud/libs/telemetry"
)

func WriteJSON(ct context.Context, dataCollector telemetry.DataCollector, writer http.ResponseWriter, body interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)

	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(buf)
}
