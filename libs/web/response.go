package web

import (
	"encoding/json"
	"net/http"

	"github.com/teamyapp/cloud/libs/obs"
)

func WriteJSON(dataCollector obs.DataCollector, writer http.ResponseWriter, body interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)

	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(buf)
}
