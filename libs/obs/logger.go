package obs

import (
	"context"
)

type Logger interface {
	Log(severity Severity, props Props)
	LogWithContext(ct context.Context, severity Severity, props Props)
	LogAndSkip(severity Severity, props Props, skipCallers int)
	LogWithContextAndSkip(ct context.Context, severity Severity, props Props, skipCallers int)
}
