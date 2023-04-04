package retry

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Infinite struct {
	dataCollector   telemetry.DataCollector
	runtime         runtime.Runtime
	shortBackOff    backoff.BackOff
	longBackOff     backoff.BackOff
	beforeSkipRetry *func()
}

var _ Retry = (*Infinite)(nil)

func (i Infinite) WithRetry(ct context.Context, execute func() *errs.Error) (int, *errs.Error) {
	var retries int
	for {
		err := execute()
		retries++
		if err == nil {
			i.shortBackOff.OnSuccess()
			i.longBackOff.OnSuccess()
			return retries, nil
		}

		i.dataCollector.Logger.ErrorWithContext(ct, err)
		category := errs.GetErrorCategory(err.Code)
		i.dataCollector.Logger.WarningWithContext(ct, fmt.Sprintf("Err category: %s", category))

		switch category {
		case errs.ClientInteraction:
			if i.beforeSkipRetry != nil {
				(*i.beforeSkipRetry)()
			}

			return retries, err
		case errs.Transient:
			i.shortBackOff.OnFailure()
			i.dataCollector.Logger.Info(fmt.Sprintf("Retry after %v", i.shortBackOff.Delay()))
			i.runtime.Sleep(i.shortBackOff.Delay())
		case errs.Outage:
			i.longBackOff.OnFailure()
			i.dataCollector.Logger.Info(fmt.Sprintf("Retry after %v", i.longBackOff.Delay()))
			i.runtime.Sleep(i.longBackOff.Delay())
		}

	}
}

func NewInfinite(
	dataCollector telemetry.DataCollector,
	runtime runtime.Runtime,
	shortBackOff backoff.BackOff,
	longBackOff backoff.BackOff,
	beforeSkipRetry *func(),
) Infinite {
	return Infinite{
		dataCollector:   dataCollector,
		runtime:         runtime,
		shortBackOff:    shortBackOff,
		longBackOff:     longBackOff,
		beforeSkipRetry: beforeSkipRetry,
	}
}
