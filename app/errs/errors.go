package errs

type NotFound struct {
	Message string
}

var _ error = (*NotFound)(nil)

func (n NotFound) Error() string {
	return n.Message
}
