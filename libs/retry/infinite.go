package retry

import (
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
)

type Infinite struct {
	backoff backoff.BackOff
	runtime runtime.Runtime
}

var _ Retry = (*Infinite)(nil)

func (i Infinite) WithRetry(execute func() error) (int, error) {
	var retries int
	for {
		err := execute()
		if err == nil {
			i.backoff.OnSuccess()
			return retries, nil
		}

		i.backoff.OnFailure()
		i.runtime.Sleep(i.backoff.Delay())
		retries++
	}
}

func NewInfinite(runtime runtime.Runtime, backoff backoff.BackOff) Infinite {
	return Infinite{runtime: runtime, backoff: backoff}
}
