package errs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

var toHTTPStatusCode = map[ErrorCode]int{
	Unknown:          http.StatusTeapot,
	InvalidArgument:  http.StatusBadRequest,
	InvalidOperation: http.StatusBadRequest,
	Timeout:          http.StatusRequestTimeout,
	NotFound:         http.StatusNotFound,
	AlreadyExists:    http.StatusConflict,
	Unauthenticated:  http.StatusUnauthorized,
	PermissionDenied: http.StatusForbidden,
	Unimplemented:    http.StatusNotImplemented,
	NotReady:         http.StatusServiceUnavailable,
}

type errorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string
}

func GetFromHTTPErr(response *http.Response) *Error {
	// Informational responses (100 – 199)
	// Successful responses (200 – 299)
	// Redirection messages (300 – 399)
	// Client error responses (400 – 499)
	// Server error responses (500 – 599)
	if response.StatusCode < 400 {
		return nil
	}

	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return &Error{
			Code:     Unknown,
			EmbedErr: err,
		}
	}

	errRes := errorResponse{}
	err = json.Unmarshal(buf, &errRes)
	if err != nil {
		return &Error{
			Code:     Unknown,
			EmbedErr: err,
		}
	}

	response.Body = io.NopCloser(bytes.NewBuffer(buf))
	return &Error{
		Code:    errRes.Code,
		Message: errRes.Message,
	}
}

func InsertHTTPErr(err *Error, responseWriter http.ResponseWriter) {
	if err == nil {
		return
	}

	errRes := errorResponse{
		Code:    err.Code,
		Message: err.Message,
	}

	buf, jsonErr := json.Marshal(errRes)
	if jsonErr != nil {
		return
	}

	statusCode, ok := toHTTPStatusCode[err.Code]
	if ok {
		responseWriter.WriteHeader(statusCode)
	} else {
		responseWriter.WriteHeader(toHTTPStatusCode[Unknown])
	}

	responseWriter.Write(buf)
}
