package telemetry

import (
	"context"
)

type Logger interface {
	Log(level LogLevel, props Props)
	LogWithContext(ct context.Context, level LogLevel, props Props)
	LogAndSkip(level LogLevel, props Props, skipCallers int)
	LogWithContextAndSkip(ct context.Context, level LogLevel, props Props, skipCallers int)
}
