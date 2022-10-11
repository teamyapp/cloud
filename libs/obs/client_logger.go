package obs

import (
	"context"

	"github.com/teamyapp/cloud/libs/ctx"
)

type ClientLogger struct {
	logger Logger
}

var _ Logger = (*ClientLogger)(nil)

func (c ClientLogger) Log(level VisibleLevel, props Props) {
	c.LogAndSkip(level, props, 1)
}

func (c ClientLogger) LogAndSkip(level VisibleLevel, props Props, skipCallers int) {
	c.logger.LogAndSkip(level, props, skipCallers+1)
}

func (c ClientLogger) LogWithContext(ct context.Context, level VisibleLevel, props Props) {
	c.LogWithContextAndSkip(ct, level, props, 1)
}

func (c ClientLogger) LogWithContextAndSkip(ct context.Context, level VisibleLevel, props Props, skipCallers int) {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	clientID, ok := ctx.GetClientID(ct)
	if ok {
		newProps["clientId"] = clientID
	}

	c.logger.LogWithContextAndSkip(ct, level, newProps, skipCallers+1)
}

func NewClientLogger(logger Logger) ClientLogger {
	return ClientLogger{
		logger: logger,
	}
}
