package obs

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

type RawLogger struct {
	visibleSeverity Severity
}

var _ Logger = (*RawLogger)(nil)

func (r RawLogger) Log(severity Severity, properties Props) {
	if severities[severity] >= severities[r.visibleSeverity] {
		addDefaultProps(severity, properties, 1)
		buf, err := json.Marshal(properties)
		if err != nil {
			return
		}

		fmt.Println(string(buf))
	}
}

func addDefaultProps(severity Severity, props Props, skipCallers int) {
	_, fileName, lineNum, ok := runtime.Caller(skipCallers + 1)
	if !ok {
		return
	}

	props[HappenAtProp] = time.Now().UTC()
	props[SeverityProp] = severity
	props[FileNameProp] = fileName
	props[LineNumberProp] = int64(lineNum)
}

func NewRawLogger(visibleSeverity Severity) RawLogger {
	return RawLogger{visibleSeverity: visibleSeverity}
}
