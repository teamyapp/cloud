package obs

import (
	"context"

	"github.com/teamyapp/cloud/libs/ctx"
)

type RequestLogger struct {
	logger Logger
	ct     context.Context
}

var _ Logger = (*RequestLogger)(nil)

func (r RequestLogger) Log(severity Severity, props Props) {
	r.LogAndSkip(severity, props, 1)
}

func (r RequestLogger) LogAndSkip(severity Severity, props Props, skipCallers int) {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	requestID := ctx.GetRequestID(r.ct)
	if len(requestID) > 0 {
		newProps["requestId"] = requestID
	}

	r.logger.LogAndSkip(severity, newProps, skipCallers+1)
}

func NewRequestLogger(logger Logger, ct context.Context) RequestLogger {
	return RequestLogger{
		logger: logger,
		ct:     ct,
	}
}
