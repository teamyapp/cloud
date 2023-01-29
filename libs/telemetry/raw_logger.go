package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

const logBufferSize = 2000

type RawLogger struct {
	visibleLevel LogLevel
	logQueue     chan Props
}

var _ Logger = (*RawLogger)(nil)

func (r RawLogger) Log(level LogLevel, properties Props) {
	r.LogAndSkip(level, properties, 1)
}

func (r RawLogger) LogAndSkip(level LogLevel, properties Props, skipCallers int) {
	if logLevelRank[level] <= logLevelRank[r.visibleLevel] {
		properties = withDefaults(level, properties, skipCallers+1)
		r.logQueue <- properties
	}
}

func (r RawLogger) LogWithContext(ct context.Context, level LogLevel, props Props) {
	r.LogWithContextAndSkip(ct, level, props, 1)
}

func (r RawLogger) LogWithContextAndSkip(ct context.Context, level LogLevel, properties Props, skipCallers int) {
	r.LogAndSkip(level, properties, skipCallers+1)
}

func withDefaults(level LogLevel, props Props, skipCallers int) Props {
	_, fileName, lineNum, ok := runtime.Caller(skipCallers + 1)
	if !ok {
		return props
	}

	newProps := Props{}
	for key, value := range props {
		if key == CauseProp {
			newProps[key] = fmt.Sprintf("%v", value)
		} else {
			newProps[key] = value
		}
	}

	newProps[HappenAtProp] = time.Now().UTC()
	newProps[SeverityProp] = level
	newProps[FileNameProp] = fileName
	newProps[LineNumberProp] = int64(lineNum)
	return newProps
}

func NewRawLogger(visibleLevel LogLevel) RawLogger {
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
		visibleLevel: visibleLevel,
		logQueue:     logQueue,
	}
}
