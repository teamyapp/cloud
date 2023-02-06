package errs

import (
	"fmt"
)

type ResolverError struct {
	code    ErrorCode
	message string
}

var _ error = (*ResolverError)(nil)

func (r ResolverError) Error() string {
	return fmt.Sprintf("code=%s, message=%s", r.code, r.message)
}

func (r ResolverError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code":    r.code,
		"message": r.message,
	}
}

func ToResolverErr(err *Error) *ResolverError {
	if err == nil {
		return nil
	}

	return &ResolverError{
		code:    err.Code,
		message: err.Message,
	}
}
