package obs

import (
	"context"

	"github.com/teamyapp/cloud/libs/ctx"
)

type RequestLogger struct {
	logger Logger
}

var _ Logger = (*RequestLogger)(nil)

func (r RequestLogger) Log(level LogLevel, props Props) {
	r.LogAndSkip(level, props, 1)
}

func (r RequestLogger) LogAndSkip(level LogLevel, props Props, skipCallers int) {
	r.logger.LogAndSkip(level, props, skipCallers+1)
}

func (r RequestLogger) LogWithContext(ct context.Context, level LogLevel, props Props) {
	r.LogWithContextAndSkip(ct, level, props, 1)
}

func (r RequestLogger) LogWithContextAndSkip(ct context.Context, level LogLevel, props Props, skipCallers int) {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	requestID, ok := ctx.GetRequestID(ct)
	if ok {
		newProps["RequestId"] = requestID
	}

	r.logger.LogWithContextAndSkip(ct, level, newProps, skipCallers+1)
}

func NewRequestLogger(logger Logger) RequestLogger {
	return RequestLogger{
		logger: logger,
	}
}
