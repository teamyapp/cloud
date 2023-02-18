package telemetry

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type frame struct {
	FunctionName string
	FileName     string
	LineNumber   int
}

func (f frame) shortFunctionName() string {
	index := strings.LastIndex(f.FunctionName, string(filepath.Separator))
	funcName := f.FunctionName[index+1:]
	index = strings.Index(funcName, ".")
	return funcName[index+1:]
}

func (f frame) String() string {
	return fmt.Sprintf("%v:%v %v", f.FileName, f.LineNumber, f.shortFunctionName())
}

type stackTrace struct {
	frames []frame
}

var _ json.Marshaler = (*stackTrace)(nil)

func (s stackTrace) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s stackTrace) String() string {
	var frameLines []string
	for _, fm := range s.frames {
		frameLines = append(frameLines, fm.String())
	}

	return "\n" + strings.Join(frameLines, "\n")
}

func newStackTrace(maxStackSize int, skipCallers int) stackTrace {
	callers := make([]uintptr, maxStackSize)

	// skip runtime.Callers and newStackTrace
	n := runtime.Callers(skipCallers+2, callers)
	callers = callers[0 : n-1]
	callersFrames := runtime.CallersFrames(callers)

	var frames []frame
	for {
		fm, hasNext := callersFrames.Next()
		if !hasNext {
			break
		}

		frames = append(frames, frame{
			FunctionName: fm.Function,
			FileName:     fm.File,
			LineNumber:   fm.Line,
		})
	}

	return stackTrace{frames: frames}
}
