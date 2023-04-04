package retry

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type MaxCount struct {
	dataCollector   telemetry.DataCollector
	runtime         runtime.Runtime
	shortBackOff    backoff.BackOff
	longBackOff     backoff.BackOff
	maxCount        int
	beforeSkipRetry *func()
}

var _ Retry = (*MaxCount)(nil)

func (m MaxCount) WithRetry(ct context.Context, execute func() *errs.Error) (int, *errs.Error) {
	var err *errs.Error
	for retry := 1; retry <= m.maxCount; retry++ {
		err = execute()
		if err == nil {
			m.shortBackOff.OnSuccess()
			m.longBackOff.OnSuccess()
			return retry, nil
		}

		m.dataCollector.Logger.ErrorWithContext(ct, err)
		category := errs.GetErrorCategory(err.Code)
		m.dataCollector.Logger.WarningWithContext(ct, fmt.Sprintf("Err category: %s", category))

		switch category {
		case errs.ClientInteraction:
			if m.beforeSkipRetry != nil {
				(*m.beforeSkipRetry)()
			}

			return retry, err
		case errs.Transient:
			m.shortBackOff.OnFailure()
			m.dataCollector.Logger.Info(fmt.Sprintf("Retry after %v", m.shortBackOff.Delay()))
			m.runtime.Sleep(m.shortBackOff.Delay())
		case errs.Outage:
			m.longBackOff.OnFailure()
			m.dataCollector.Logger.Info(fmt.Sprintf("Retry after %v", m.longBackOff.Delay()))
			m.runtime.Sleep(m.longBackOff.Delay())
		}
	}

	return m.maxCount, err
}

func NewMaxCount(
	dataCollector telemetry.DataCollector,
	runtime runtime.Runtime,
	shortBackOff backoff.BackOff,
	longBackOff backoff.BackOff,
	maxCount int,
	beforeSkipRetry *func(),
) MaxCount {
	return MaxCount{
		dataCollector:   dataCollector,
		runtime:         runtime,
		shortBackOff:    shortBackOff,
		longBackOff:     longBackOff,
		maxCount:        maxCount,
		beforeSkipRetry: beforeSkipRetry,
	}
}
