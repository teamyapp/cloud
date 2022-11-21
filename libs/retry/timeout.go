package retry

import (
	"fmt"
	"time"

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

func (t Timeout) WithRetry(execute func() error) (int, error) {
	timeoutAt := t.clock.Now().Add(t.timeout)
	var retryCount int
	var err error
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
		if !t.clock.Now().Add(t.backoff.Delay()).Before(timeoutAt) {
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
	_backOff backoff.BackOff,
	clock runtime.Clock,
	timeout time.Duration,
	beforeRetryDelay func()) Timeout {
	return Timeout{
		runtime:          runtime,
		backoff:          _backOff,
		clock:            clock,
		timeout:          timeout,
		beforeRetryDelay: beforeRetryDelay,
	}
}
