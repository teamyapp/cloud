package errs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const HTTPClientErrors = 400
const HTTPServerErrors = 500

var toHTTPStatusCode = map[ErrorCode]int{
	Unknown:          http.StatusInternalServerError,
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
	Message string    `json:"message"`
}

func GetFromHTTPErr(response *http.Response) *Error {
	// Informational responses (100 – 199)
	// Successful responses (200 – 299)
	// Redirection messages (300 – 399)
	// Client error responses (400 – 499)
	// Server error responses (500 – 599)
	if response.StatusCode < HTTPClientErrors {
		return nil
	}

	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return NewErrorSkipCallers(Unknown, err.Error(), 1)
	}

	errRes := errorResponse{}
	err = json.Unmarshal(buf, &errRes)
	if err != nil {
		return NewErrorSkipCallers(Unknown, err.Error(), 1)
	}

	response.Body = io.NopCloser(bytes.NewBuffer(buf))
	return NewErrorSkipCallers(errRes.Code, errRes.Message, 1)
}

func SetHTTPErr(err *Error, responseWriter http.ResponseWriter) {
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
