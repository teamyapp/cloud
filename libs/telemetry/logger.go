package telemetry

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	tmio "github.com/teamyapp/cloud/libs/io"
)

const maxStackDepth = 100

type Logger struct {
	visibleLevel    LogLevel
	logInterceptors []LogInterceptor
	lineFormatter   LineFormatter
	output          io.Writer
}

func (l Logger) Fatal(err *errs.Error) {
	props := Props{
		CauseProp:      err,
		StackTraceProp: newStackTrace(maxStackDepth, 1),
	}
	l.log(Fatal, props, 1)
	panic(err)
}

func (l Logger) FatalWithContext(ct context.Context, err *errs.Error) {
	props := Props{
		CauseProp:      err,
		StackTraceProp: newStackTrace(maxStackDepth, 1),
	}
	l.logWithContext(ct, Fatal, props, 1)
	panic(err)
}

func (l Logger) Error(err *errs.Error) {
	props := Props{
		CauseProp:      err,
		StackTraceProp: newStackTrace(maxStackDepth, 1),
	}
	l.log(Error, props, 1)
}

func (l Logger) ErrorWithContext(ct context.Context, err *errs.Error) {
	props := Props{
		CauseProp:      err,
		StackTraceProp: newStackTrace(maxStackDepth, 1),
	}
	l.logWithContext(ct, Error, props, 1)
}

func (l Logger) Warning(input interface{}) {
	l.log(Warning, Props{MessageProp: input}, 1)
}

func (l Logger) WarningWithContext(ct context.Context, input interface{}) {
	l.logWithContext(ct, Warning, Props{MessageProp: input}, 1)
}

func (l Logger) Info(input interface{}) {
	l.log(Info, Props{MessageProp: input}, 1)
}

func (l Logger) InfoWithContext(ct context.Context, input interface{}) {
	l.logWithContext(ct, Info, Props{MessageProp: input}, 1)
}

func (l Logger) Debug(input interface{}) {
	l.log(Debug, Props{MessageProp: input}, 1)
}

func (l Logger) DebugWithContext(ct context.Context, input interface{}) {
	l.logWithContext(ct, Debug, Props{MessageProp: input}, 1)
}

func (l Logger) Log(level LogLevel, props Props) {
	l.log(level, props, 1)
}

func (l Logger) LogWithContext(ct context.Context, level LogLevel, props Props) {
	l.logWithContext(ct, level, props, 1)
}

func (l Logger) log(level LogLevel, props Props, skipCallers int) {
	l.logWithContext(context.Background(), level, props, skipCallers+1)
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

func NewLogOutput(environment env.Environment, logFilePath string) (io.WriteCloser, *errs.Error) {
	if environment == env.DevelopmentEnv {
		logDir := filepath.Dir(logFilePath)

		// MkdirAll requires at least 700 permission:
		// https://github.com/golang/go/issues/22323
		err := os.MkdirAll(logDir, 0744)
		if err != nil {
			return nil, &errs.Error{
				Code:     errs.OS,
				EmbedErr: err,
			}
		}

		file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			return nil, &errs.Error{
				Code:     errs.OS,
				EmbedErr: err,
			}
		}

		return tmio.NewMultiWriteCloser(file, os.Stdout), nil
	}

	return os.Stdout, nil
}
