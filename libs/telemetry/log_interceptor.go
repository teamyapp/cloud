package telemetry

import (
	"context"

	"github.com/teamyapp/cloud/libs/ctx"
)

type LogInterceptor func(ct context.Context, level LogLevel, props Props) Props

func NewCommitLogInterceptor(commit string) LogInterceptor {
	return func(ct context.Context, level LogLevel, props Props) Props {
		newProps := Props{}
		for key, value := range props {
			newProps[key] = value
		}

		newProps[CommitProp] = commit
		return newProps
	}
}

func NewServiceLogInterceptor(serviceName string) LogInterceptor {
	return func(ct context.Context, level LogLevel, props Props) Props {
		newProps := Props{}
		for key, value := range props {
			newProps[key] = value
		}

		newProps[ServiceNameProp] = serviceName
		return newProps
	}
}

func RequestLogInterceptor(ct context.Context, level LogLevel, props Props) Props {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	requestID, ok := ctx.GetRequestID(ct)
	if ok {
		newProps[RequestIDProp] = requestID
	}

	return newProps
}

var _ LogInterceptor = RequestLogInterceptor

func ClientLogInterceptor(ct context.Context, level LogLevel, props Props) Props {
	newProps := Props{}
	for key, value := range props {
		newProps[key] = value
	}

	clientID, ok := ctx.GetClientID(ct)
	if ok {
		newProps[ClientIDProp] = clientID
	}

	return newProps
}

var _ LogInterceptor = ClientLogInterceptor
