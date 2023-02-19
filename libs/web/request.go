package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

func WriteJSONToRequest(ct context.Context, dataCollector telemetry.DataCollector, req *http.Request, body interface{}) *errs.Error {
	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewBuffer(buf))
	return nil
}
