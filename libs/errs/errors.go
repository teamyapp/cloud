package errs

import (
	"fmt"
)

type ErrorCode string

const (
	Unknown           ErrorCode = "Unknown"
	Cancelled         ErrorCode = "Cancelled"
	InvalidArgument   ErrorCode = "InvalidArgument"
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
	Serialization     ErrorCode = "Serialization"
	Deserialization   ErrorCode = "Deserialization"
)

type Error struct {
	Code     ErrorCode
	EmbedErr error
	Message  string
}

var _ error = (*Error)(nil)

func (e Error) Error() string {
	return fmt.Sprintf("[Error Code=%v Message=%v EmbedErr=%v]", e.Code, e.Message, e.EmbedErr)
}
