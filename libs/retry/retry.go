package retry

type Retry interface {
	WithRetry(execute func() error) (int, error)
}
