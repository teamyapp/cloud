package obs

import (
	"context"
)

type Logger interface {
	Log(level VisibleLevel, props Props)
	LogWithContext(ct context.Context, level VisibleLevel, props Props)
	LogAndSkip(level VisibleLevel, props Props, skipCallers int)
	LogWithContextAndSkip(ct context.Context, level VisibleLevel, props Props, skipCallers int)
}
