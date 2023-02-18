package errs

import (
	"encoding/json"
	"fmt"
)

type ErrorCode string

const (
	Unknown           ErrorCode = "Unknown"
	Cancelled         ErrorCode = "Cancelled"
	InvalidArgument   ErrorCode = "InvalidArgument"
	InvalidValue      ErrorCode = "InvalidValue"
	InvalidFormat     ErrorCode = "InvalidFormat"
	InvalidOperation  ErrorCode = "InvalidOperation"
	Aborted           ErrorCode = "Aborted"
	Timeout           ErrorCode = "Timeout"
	NotFound          ErrorCode = "NotFound"
	AlreadyExists     ErrorCode = "AlreadyExist"
	Unauthenticated   ErrorCode = "Unauthenticated"
	PermissionDenied  ErrorCode = "PermissionDenied"
	ResourceExhausted ErrorCode = "ResourceExhausted"
	Unimplemented     ErrorCode = "Unimplemented"
	NotReady          ErrorCode = "NotReady"
	Unreachable       ErrorCode = "Unreachable"
	IO                ErrorCode = "IO"
	OS                ErrorCode = "OS"
	Serialization     ErrorCode = "Serialization"
	Deserialization   ErrorCode = "Deserialization"
)

type Error struct {
	Code     ErrorCode
	EmbedErr error
	Message  string
}

var _ json.Marshaler = (*Error)(nil)

func (e Error) MarshalJSON() ([]byte, error) {
	fields := map[string]string{
		"Code":    string(e.Code),
		"Message": e.Message,
	}

	if e.EmbedErr != nil {
		fields["EmbedErr"] = e.EmbedErr.Error()
	}

	return json.Marshal(fields)
}

func (e Error) String() string {
	return fmt.Sprintf("[Code=%v Message=%v EmbedErr=%v]", e.Code, e.Message, formatErr(e.EmbedErr))
}

func (e Error) ToError() error {
	return fmt.Errorf("[Code=%v Message=%v EmbedErr=%v]", e.Code, e.Message, formatErr(e.EmbedErr))
}

func formatErr(err error) string {
	if err == nil {
		return "nil"
	}

	return err.Error()
}
