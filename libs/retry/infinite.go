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
	logger          telemetry.Logger
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

		i.logger.ErrorWithContext(ct, err)
		category := errs.GetErrorCategory(err.Code)
		i.logger.WarningWithContext(ct, fmt.Sprintf("Err category: %s", category))

		switch category {
		case errs.ClientInteraction:
			if i.beforeSkipRetry != nil {
				(*i.beforeSkipRetry)()
			}

			return retries, err
		case errs.Transient:
			i.shortBackOff.OnFailure()
			i.logger.Info(fmt.Sprintf("Retry after %v", i.shortBackOff.Delay()))
			i.runtime.Sleep(i.shortBackOff.Delay())
		case errs.Outage:
			i.longBackOff.OnFailure()
			i.logger.Info(fmt.Sprintf("Retry after %v", i.longBackOff.Delay()))
			i.runtime.Sleep(i.longBackOff.Delay())
		}

	}
}

func NewInfinite(
	logger telemetry.Logger,
	runtime runtime.Runtime,
	shortBackOff backoff.BackOff,
	longBackOff backoff.BackOff,
	beforeSkipRetry *func(),
) Infinite {
	return Infinite{
		logger:          logger,
		runtime:         runtime,
		shortBackOff:    shortBackOff,
		longBackOff:     longBackOff,
		beforeSkipRetry: beforeSkipRetry,
	}
}
