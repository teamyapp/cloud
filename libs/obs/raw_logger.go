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
	visibleLevel VisibleLevel
	logQueue     chan Props
}

var _ Logger = (*RawLogger)(nil)

func (r RawLogger) Log(level VisibleLevel, properties Props) {
	r.LogAndSkip(level, properties, 1)
}

func (r RawLogger) LogAndSkip(level VisibleLevel, properties Props, skipCallers int) {
	if visibleLevelRank[level] <= visibleLevelRank[r.visibleLevel] {
		properties = withDefaults(level, properties, skipCallers+1)
		r.logQueue <- properties
	}
}

func (r RawLogger) LogWithContext(ct context.Context, level VisibleLevel, props Props) {
	r.LogWithContextAndSkip(ct, level, props, 1)
}

func (r RawLogger) LogWithContextAndSkip(ct context.Context, level VisibleLevel, properties Props, skipCallers int) {
	r.LogAndSkip(level, properties, skipCallers+1)
}

func withDefaults(level VisibleLevel, props Props, skipCallers int) Props {
	_, fileName, lineNum, ok := runtime.Caller(skipCallers + 1)
	if !ok {
		return props
	}

	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	newProps[HappenAtProp] = time.Now().UTC()
	newProps[SeverityProp] = level
	newProps[FileNameProp] = fileName
	newProps[LineNumberProp] = int64(lineNum)
	return newProps
}

func NewRawLogger(visibleLevel VisibleLevel) RawLogger {
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
