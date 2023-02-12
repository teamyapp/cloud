package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

func WriteJSON(ct context.Context, dataCollector telemetry.DataCollector, writer http.ResponseWriter, body interface{}) {
	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	writer.Write(buf)
}

type Client func(ct context.Context, req *http.Request) (*http.Response, error)

func (c *Client) Do(ct context.Context, req *http.Request) (*http.Response, error) {
	return (*c)(ct, req)
}
