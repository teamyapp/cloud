package obs

import (
	"context"

	"github.com/teamyapp/cloud/libs/ctx"
)

type ClientLogger struct {
	logger Logger
}

var _ Logger = (*ClientLogger)(nil)

func (c ClientLogger) Log(severity Severity, props Props) {
	c.LogAndSkip(severity, props, 1)
}

func (c ClientLogger) LogAndSkip(severity Severity, props Props, skipCallers int) {
	c.logger.LogAndSkip(severity, props, skipCallers+1)
}

func (c ClientLogger) LogWithContext(ct context.Context, severity Severity, props Props) {
	c.LogWithContextAndSkip(ct, severity, props, 1)
}

func (c ClientLogger) LogWithContextAndSkip(ct context.Context, severity Severity, props Props, skipCallers int) {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	clientID, ok := ctx.GetClientID(ct)
	if ok {
		newProps["clientId"] = clientID
	}

	c.logger.LogWithContextAndSkip(ct, severity, newProps, skipCallers+1)
}

func NewClientLogger(logger Logger) ClientLogger {
	return ClientLogger{
		logger: logger,
	}
}
