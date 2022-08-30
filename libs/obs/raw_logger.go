package obs

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

type RawLogger struct {
	visibleLogLevel LogLevel
}

var _ Logger = (*RawLogger)(nil)

func (r RawLogger) Log(logLevel LogLevel, properties Props) {
	if severity[logLevel] >= severity[r.visibleLogLevel] {
		addDefaultProps(logLevel, properties, 1)
		buf, err := json.Marshal(properties)
		if err != nil {
			return
		}

		fmt.Println(string(buf))
	}
}

func addDefaultProps(logLevel LogLevel, props Props, skipCallers int) {
	_, fileName, lineNum, ok := runtime.Caller(skipCallers + 1)
	if !ok {
		return
	}

	props["happenAt"] = time.Now().UTC()
	props["logLevel"] = logLevel
	props["fileName"] = fileName
	props["lineNumber"] = int64(lineNum)
}

func NewRawLogger(visibleLogLevel LogLevel) RawLogger {
	return RawLogger{visibleLogLevel: visibleLogLevel}
}
