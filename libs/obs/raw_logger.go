package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

const logBufferSize = 2000

type RawLogger struct {
	visibleSeverity Severity
	logQueue        chan Props
}

var _ Logger = (*RawLogger)(nil)

func (r RawLogger) Log(severity Severity, properties Props) {
	r.LogAndSkip(severity, properties, 1)
}

func (r RawLogger) LogAndSkip(severity Severity, properties Props, skipCallers int) {
	if severities[severity] >= severities[r.visibleSeverity] {
		properties = withDefaults(severity, properties, skipCallers+1)
		r.logQueue <- properties
	}
}

func (r RawLogger) LogWithContext(ct context.Context, severity Severity, props Props) {
	r.LogWithContextAndSkip(ct, severity, props, 1)
}

func (r RawLogger) LogWithContextAndSkip(ct context.Context, severity Severity, properties Props, skipCallers int) {
	r.LogAndSkip(severity, properties, skipCallers+1)
}

func withDefaults(severity Severity, props Props, skipCallers int) Props {
	_, fileName, lineNum, ok := runtime.Caller(skipCallers + 1)
	if !ok {
		return props
	}

	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	newProps[HappenAtProp] = time.Now().UTC()
	newProps[SeverityProp] = severity
	newProps[FileNameProp] = fileName
	newProps[LineNumberProp] = int64(lineNum)
	return newProps
}

func NewRawLogger(visibleSeverity Severity) RawLogger {
	logQueue := make(chan Props, logBufferSize)
	go func() {
		for logProps := range logQueue {
			buf, err := json.Marshal(logProps)
			if err != nil {
				return
			}

			fmt.Println(string(buf))
		}
	}()
	return RawLogger{
		visibleSeverity: visibleSeverity,
		logQueue:        logQueue,
	}
}
