package telemetry

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
)

type Logger struct {
	visibleLevel    LogLevel
	logInterceptors []LogInterceptor
	lineFormatter   LineFormatter
	output          io.Writer
}

func (l Logger) Fatal(err *errs.Error) {
	l.FatalWithContext(context.Background(), err)
	panic(err)
}

func (l Logger) FatalWithContext(ct context.Context, err *errs.Error) {
	l.logWithContext(ct, Fatal, Props{CauseProp: err}, 1)
}

func (l Logger) Error(err *errs.Error) {
	l.ErrorWithContext(context.Background(), err)
}

func (l Logger) ErrorWithContext(ct context.Context, err *errs.Error) {
	l.logWithContext(ct, Error, Props{CauseProp: err}, 1)
}

func (l Logger) Warning(input interface{}) {
	l.WarningWithContext(context.Background(), input)
}

func (l Logger) WarningWithContext(ct context.Context, input interface{}) {
	l.logWithContext(ct, Warning, Props{MessageProp: input}, 1)
}

func (l Logger) Info(input interface{}) {
	l.InfoWithContext(context.Background(), input)
}

func (l Logger) InfoWithContext(ct context.Context, input interface{}) {
	l.logWithContext(ct, Info, Props{MessageProp: input}, 1)
}

func (l Logger) Debug(input interface{}) {
	l.DebugWithContext(context.Background(), input)
}

func (l Logger) DebugWithContext(ct context.Context, input interface{}) {
	l.logWithContext(ct, Debug, Props{MessageProp: input}, 1)
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
		newProps[key] = value
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
