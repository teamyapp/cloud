package telemetry

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"
)

type Logger struct {
	visibleLevel    LogLevel
	logInterceptors []LogInterceptor
	lineFormatter   LineFormatter
	output          io.Writer
}

func (l Logger) Log(level LogLevel, props Props) {
	l.logWithContext(context.Background(), level, props, 1)
}

func (l Logger) LogWithContext(ct context.Context, level LogLevel, props Props) {
	l.logWithContext(ct, level, props, 1)
}

func (l Logger) logWithContext(ct context.Context, level LogLevel, props Props, skipCallers int) {
	if logLevelRank[level] > logLevelRank[l.visibleLevel] {
		return
	}

	for _, interceptor := range l.logInterceptors {
		props = interceptor(ct, level, props)
	}

	props = withDefaults(level, props, skipCallers+1)
	line, err := l.lineFormatter.FormatLine(props)
	if err != nil {
		return
	}

	_, _ = fmt.Fprintln(l.output, line)
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

func NewLogger(
	lineFormatter LineFormatter,
	output io.Writer,
	visibleLevel LogLevel,
	logInterceptors []LogInterceptor,
) Logger {
	return Logger{
		lineFormatter:   lineFormatter,
		output:          output,
		visibleLevel:    visibleLevel,
		logInterceptors: logInterceptors,
	}
}
