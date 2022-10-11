package obs

import (
	"context"
)

type ServiceLogger struct {
	serviceName string
	logger      Logger
}

var _ Logger = (*ServiceLogger)(nil)

func (s ServiceLogger) Log(level VisibleLevel, props Props) {
	s.LogAndSkip(level, props, 1)
}

func (s ServiceLogger) LogAndSkip(level VisibleLevel, props Props, skipCallers int) {
	s.logger.LogAndSkip(level, s.withServiceProps(props), skipCallers+1)
}

func (s ServiceLogger) LogWithContext(ct context.Context, level VisibleLevel, props Props) {
	s.LogWithContextAndSkip(ct, level, props, 1)
}

func (s ServiceLogger) LogWithContextAndSkip(ct context.Context, level VisibleLevel, props Props, skipCallers int) {
	s.logger.LogWithContextAndSkip(ct, level, s.withServiceProps(props), skipCallers+1)
}

func (s ServiceLogger) withServiceProps(props Props) Props {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	newProps["serviceName"] = s.serviceName
	return newProps
}

func NewServiceLogger(serviceName string, logger Logger) ServiceLogger {
	return ServiceLogger{
		serviceName: serviceName,
		logger:      logger,
	}
}
