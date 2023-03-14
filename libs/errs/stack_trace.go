package errs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const maxStackDepth = 100

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

type StackTrace struct {
	frames []frame
}

var _ json.Marshaler = (*StackTrace)(nil)

func (s StackTrace) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s StackTrace) String() string {
	var frameLines []string
	for _, fm := range s.frames {
		frameLines = append(frameLines, fm.String())
	}

	return strings.Join(frameLines, "\n")
}

func newStackTrace(maxStackSize int, skipCallers int) StackTrace {
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

	return StackTrace{frames: frames}
}
