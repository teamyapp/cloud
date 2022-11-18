package retry

import (
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
)

type MaxCount struct {
	runtime  runtime.Runtime
	backoff  backoff.BackOff
	maxCount int
}

var _ Retry = (*MaxCount)(nil)

func (m MaxCount) WithRetry(execute func() error) (int, error) {
	var err error
	var retryCount int
	for retry := 0; retry < m.maxCount; retry++ {
		err = execute()
		if err == nil {
			m.backoff.OnSuccess()
			return retryCount, nil
		}

		m.backoff.OnFailure()
		m.runtime.Sleep(m.backoff.Delay())
	}

	return m.maxCount, err
}

func NewMaxCount(runtime runtime.Runtime, backoff backoff.BackOff, maxCount inT) MaxCount {
	return MaxCount{
	    runtime: runtime, 
	    backoff: backoff, 
	    maxCount: maxCount,
	}
}
