package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
)

func WriteJSONToRequest(ct context.Context, req *http.Request, body interface{}) *errs.Error {
	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return errs.NewError(errs.Serialization, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewBuffer(buf))
	return nil
}
