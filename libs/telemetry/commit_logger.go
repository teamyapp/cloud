package telemetry

import (
	"context"
)

type CommitLogger struct {
	commit string
	logger Logger
}

var _ Logger = (*CommitLogger)(nil)

func (c CommitLogger) Log(level LogLevel, props Props) {
	c.LogAndSkip(level, props, 1)
}

func (c CommitLogger) LogAndSkip(level LogLevel, props Props, skipCallers int) {
	c.logger.LogAndSkip(level, c.withServiceProps(props), skipCallers+1)
}

func (c CommitLogger) LogWithContext(ct context.Context, level LogLevel, props Props) {
	c.LogWithContextAndSkip(ct, level, props, 1)
}

func (c CommitLogger) LogWithContextAndSkip(ct context.Context, level LogLevel, props Props, skipCallers int) {
	c.logger.LogWithContextAndSkip(ct, level, c.withServiceProps(props), skipCallers+1)
}

func (c CommitLogger) withServiceProps(props Props) Props {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	newProps["Commit"] = c.commit
	return newProps
}

func NewCommitLogger(commit string, logger Logger) CommitLogger {
	return CommitLogger{
		commit: commit,
		logger: logger,
	}
}
