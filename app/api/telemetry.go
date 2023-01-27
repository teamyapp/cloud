package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/runner"
)

type Telemetry struct {
	dataCollector telemetry.DataCollector
}

var _ runner.Service = (*Telemetry)(nil)

func (t *Telemetry) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(telemetryPathPrefix, "upload-log"),
			Method:      http.MethodPost,
			HandlerFunc: t.uploadLog,
		},
	})

	return nil
}

func (t *Telemetry) uploadLog(w http.ResponseWriter, r *http.Request) {
	ct := r.Context()
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logEntries := []string{}
	err = json.Unmarshal(buf, &logEntries)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, logEntry := range logEntries {
		fmt.Println(logEntry)
	}

	w.WriteHeader(http.StatusNoContent)
}

func NewTelemetry(dataCollector telemetry.DataCollector) *Telemetry {
	return &Telemetry{
		dataCollector: dataCollector,
	}
}
