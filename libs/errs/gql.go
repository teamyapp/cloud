package errs

type ResolverError struct {
	err *Error
}

var _ error = (*ResolverError)(nil)

func (r ResolverError) Error() string {
	return r.err.Message
}

func (r ResolverError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code":    r.err.Code,
		"message": r.err.Message,
	}
}

func ToResolverErr(err *Error) *ResolverError {
	if err == nil {
		return nil
	}

	return &ResolverError{
		err: err,
	}
}
