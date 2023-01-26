package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/teamyapp/cloud/libs/obs"
)

func WriteJSON(ct context.Context, dataCollector obs.DataCollector, writer http.ResponseWriter, body interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)

	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(buf)
}
