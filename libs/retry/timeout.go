package retry

import (
	"time"

	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
)

type Timeout struct {
	backoff          backoff.BackOff
	timeout          time.Duration
	clock            runtime.Clock
	beforeRetryDelay func()
	runtime          runtime.Runtime
}

var _ Retry = (*Timeout)(nil)

func (t Timeout) WithRetry(execute func() error) (int, error) {
	timeoutAt := t.clock.Now().Add(t.timeout)
	var retryCount int
	var err error
	for {
		err = execute()
		if t.beforeRetryDelay != nil {
			t.beforeRetryDelay()
		}

		if err == nil {
			t.backoff.OnSuccess()
			return retryCount, nil
		}

		t.backoff.OnFailure()
		if !t.clock.Now().Before(timeoutAt) {
			retryCount++
			break
		}

		t.runtime.Sleep(t.backoff.Delay())
		retryCount++
	}

	return retryCount, err
}

func NewTimeout(backoff backoff.BackOff, timeout time.Duration, clock runtime.Clock, beforeRetryDelay func(), runtime runtime.Runtime) Timeout {
	return Timeout{backoff: backoff, timeout: timeout, clock: clock, beforeRetryDelay: beforeRetryDelay, runtime: runtime}
}
