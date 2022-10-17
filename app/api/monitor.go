package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/runner"
)

type Monitor struct {
	dataCollector obs.DataCollector
}

var _ runner.Service = (*Monitor)(nil)

func (m *Monitor) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(monitorPathPrefix, "upload-log"),
			Method:      http.MethodPost,
			HandlerFunc: m.uploadLog,
		},
	})

	return nil
}

func (m *Monitor) uploadLog(w http.ResponseWriter, r *http.Request) {
	ct := r.Context()
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logEntries := []string{}
	err = json.Unmarshal(buf, &logEntries)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, logEntry := range logEntries {
		fmt.Println(logEntry)
	}

	w.WriteHeader(http.StatusNoContent)
}

func NewMonitor(dataCollector obs.DataCollector) *Monitor {
	return &Monitor{
		dataCollector: dataCollector,
	}
}
