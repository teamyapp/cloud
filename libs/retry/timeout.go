package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Timeout struct {
	dataCollector    telemetry.DataCollector
	runtime          runtime.Runtime
	shortBackOff     backoff.BackOff
	longBackOff      backoff.BackOff
	clock            runtime.Clock
	timeout          time.Duration
	beforeRetryDelay *func()
	beforeSkipRetry  *func()
}

var _ Retry = (*Timeout)(nil)

func (t Timeout) WithRetry(ct context.Context, execute func() *errs.Error) (int, *errs.Error) {
	timeoutAt := t.clock.Now().Add(t.timeout)
	var retryCount int
	var err *errs.Error
	for {
		err = execute()
		retryCount++
		if err == nil {
			t.shortBackOff.OnSuccess()
			t.longBackOff.OnSuccess()
			break
		}

		if t.beforeRetryDelay != nil {
			(*t.beforeRetryDelay)()
		}

		t.dataCollector.Logger.ErrorWithContext(ct, err)
		category := errs.GetErrorCategory(err.Code)
		t.dataCollector.Logger.WarningWithContext(ct, fmt.Sprintf("Err category: %s", category))

		switch category {
		case errs.ClientInteraction:
			if t.beforeSkipRetry != nil {
				(*t.beforeSkipRetry)()
			}

			return retryCount, err
		case errs.Transient:
			t.shortBackOff.OnFailure()
			delay := t.shortBackOff.Delay()
			expectTime := t.clock.Now().Add(delay)
			if !expectTime.Before(timeoutAt) {
				return retryCount, err
			}

			t.runtime.Sleep(delay)
		case errs.Outage:
			t.longBackOff.OnFailure()
			delay := t.longBackOff.Delay()
			expectTime := t.clock.Now().Add(delay)
			if !expectTime.Before(timeoutAt) {
				return retryCount, err
			}

			t.runtime.Sleep(delay)
		}
	}

	return retryCount, err
}

func NewTimeout(
	dataCollector telemetry.DataCollector,
	runtime runtime.Runtime,
	shortBackOff backoff.BackOff,
	longBackOff backoff.BackOff,
	clock runtime.Clock,
	timeout time.Duration,
	beforeRetryDelay *func(),
	beforeSkipRetry *func()) Timeout {
	return Timeout{
		dataCollector:    dataCollector,
		shortBackOff:     shortBackOff,
		longBackOff:      longBackOff,
		runtime:          runtime,
		clock:            clock,
		timeout:          timeout,
		beforeRetryDelay: beforeRetryDelay,
		beforeSkipRetry:  beforeSkipRetry,
	}
}
