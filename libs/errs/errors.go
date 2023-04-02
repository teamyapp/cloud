package errs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/teamyapp/cloud/libs/collect"
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

// Deprecated: Use NewError instead.
type Error struct {
	Code ErrorCode
	// Deprecated: EmbedErr is duplicate with Message and should not be used anymore.
	EmbedErr   error
	Message    string
	stackTrace *StackTrace
}

var _ json.Marshaler = (*Error)(nil)

func NewError(code ErrorCode, message string) *Error {
	stackTrace := newStackTrace(maxStackDepth, 1)
	return &Error{
		Code:       code,
		Message:    message,
		stackTrace: &stackTrace,
	}
}

func NewErrorSkipCallers(code ErrorCode, message string, skipCallers int) *Error {
	stackTrace := newStackTrace(maxStackDepth, skipCallers+1)
	return &Error{
		Code:       code,
		Message:    message,
		stackTrace: &stackTrace,
	}
}

func (e Error) MarshalJSON() ([]byte, error) {
	fields := map[string]interface{}{
		"Code":       e.Code,
		"Message":    e.Message,
		"EmbedErr":   formatErr(e.EmbedErr),
		"StackTrace": e.stackTrace,
	}

	return json.Marshal(fields)
}

func (e Error) StackTrace() *StackTrace {
	return e.stackTrace
}

func (e Error) String() string {
	return fmt.Sprintf("[Code=%v Message=%v EmbedErr=%v, StackTrace=%v]", e.Code, e.Message, formatErr(e.EmbedErr), e.stackTrace)
}

func (e Error) ToError() error {
	return fmt.Errorf("[Code=%v Message=%v EmbedErr=%v, StackTrace=%v]", e.Code, e.Message, formatErr(e.EmbedErr), e.stackTrace)
}

func MergeErrs(errs []*Error) *Error {
	// TODO(szheng2207): find a better way to handle multiple errors instead
	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return &Error{
		Code: Unknown,
		Message: strings.Join(collect.Map(errs, func(err *Error, _ int) string {
			return err.String()
		}), "\n"),
	}
}

func formatErr(err error) string {
	if err == nil {
		return "nil"
	}

	return err.Error()
}
