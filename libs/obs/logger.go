package obs

type Logger interface {
	Log(severity Severity, props Props)
	LogAndSkip(severity Severity, props Props, skipCallers int)
}
