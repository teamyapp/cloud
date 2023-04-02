package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
)

func WriteJSONToRequest(req *http.Request, body interface{}) *errs.Error {
	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return errs.NewError(errs.Serialization, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewBuffer(buf))
	return nil
}
