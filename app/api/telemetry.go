package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Telemetry struct {
	logger telemetry.Logger
}

var _ runner.Service = (*Telemetry)(nil)

func (t *Telemetry) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodPost,
			Pattern:     path.Join(telemetryPathPrefix, "upload-log"),
			HandlerFunc: t.uploadLog,
		},
	})

	return nil
}

func (t *Telemetry) uploadLog(w http.ResponseWriter, r *http.Request) {
	ct := r.Context()
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		internalErr := errs.NewError(errs.IO, err.Error())
		t.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, w)
		return
	}

	var logEntries []string
	err = json.Unmarshal(buf, &logEntries)
	if err != nil {
		internalErr := errs.NewError(errs.Deserialization, err.Error())
		t.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, w)
		return
	}

	for _, logEntry := range logEntries {
		fmt.Println(logEntry)
	}

	w.WriteHeader(http.StatusNoContent)
}

func NewTelemetry(logger telemetry.Logger) *Telemetry {
	return &Telemetry{
		logger: logger,
	}
}
