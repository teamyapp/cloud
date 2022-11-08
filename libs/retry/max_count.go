package retry

import (
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
)

type MaxCount struct {
	backoff  backoff.BackOff
	maxCount int
	runtime  runtime.Runtime
}

var _ Retry = (*MaxCount)(nil)

func (m MaxCount) WithRetry(execute func() error) (int, error) {
	var err error
	var retryCount int
	for retryCount < m.maxCount {
		err = execute()
		if err == nil {
			m.backoff.OnSuccess()
			return retryCount, nil
		}

		m.backoff.OnFailure()
		m.runtime.Sleep(m.backoff.Delay())
		retryCount++
	}

	return retryCount, err
}

func NewMaxCount(backoff backoff.BackOff, maxCount int, runtime runtime.Runtime) MaxCount {
	return MaxCount{backoff: backoff, maxCount: maxCount, runtime: runtime}
}
