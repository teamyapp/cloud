package retry

import (
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
)

type Timeout struct {
	runtime          runtime.Runtime
	clock            runtime.Clock
	backoff          backoff.BackOff
	timeout          time.Duration
	beforeRetryDelay func()
}

var _ Retry = (*Timeout)(nil)

func (t Timeout) WithRetry(execute func() *errs.Error) (int, *errs.Error) {
	timeoutAt := t.clock.Now().Add(t.timeout)
	var retryCount int
	var err *errs.Error
	for {
		err = execute()
		if err == nil {
			t.backoff.OnSuccess()
			break
		}

		if t.beforeRetryDelay != nil {
			t.beforeRetryDelay()
		}

		t.backoff.OnFailure()
		expectTime := t.clock.Now().Add(t.backoff.Delay())
		if !expectTime.Before(timeoutAt) {
			retryCount++
			break
		}

		t.runtime.Sleep(t.backoff.Delay())
		retryCount++
	}

	return retryCount, err
}

func NewTimeout(
	runtime runtime.Runtime,
	backoff backoff.BackOff,
	clock runtime.Clock,
	timeout time.Duration,
	beforeRetryDelay func()) Timeout {
	return Timeout{
		runtime:          runtime,
		backoff:          backoff,
		clock:            clock,
		timeout:          timeout,
		beforeRetryDelay: beforeRetryDelay,
	}
}
