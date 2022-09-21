package obs

import (
	"context"
)

type ServiceLogger struct {
	serviceName string
	logger      Logger
}

var _ Logger = (*ServiceLogger)(nil)

func (s ServiceLogger) Log(severity Severity, props Props) {
	s.LogAndSkip(severity, props, 1)
}

func (s ServiceLogger) LogAndSkip(severity Severity, props Props, skipCallers int) {
	s.logger.LogAndSkip(severity, s.withServiceProps(props), skipCallers+1)
}

func (s ServiceLogger) LogWithContext(ct context.Context, severity Severity, props Props) {
	s.LogWithContextAndSkip(ct, severity, props, 1)
}

func (s ServiceLogger) LogWithContextAndSkip(ct context.Context, severity Severity, props Props, skipCallers int) {
	s.logger.LogWithContextAndSkip(ct, severity, s.withServiceProps(props), skipCallers+1)
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
