package web

import (
	"encoding/json"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
}

func WriteJSONToResponse(writer http.ResponseWriter, body interface{}) *errs.Error {
	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return errs.NewError(errs.Serialization, err.Error())
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	writer.Write(buf)
	return nil
}
