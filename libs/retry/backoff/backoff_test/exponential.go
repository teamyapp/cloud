package backoff_test

import (
	"time"

	"github.com/teamyapp/cloud/libs/retry/backoff"
)

type Exponential struct {
	nextDelay time.Duration
}

var _ backoff.BackOff = (*Exponential)(nil)

func (e *Exponential) OnSuccess() {
}

func (e *Exponential) OnFailure() {
	e.nextDelay = e.nextDelay * time.Duration(2)
}

func (e *Exponential) Delay() time.Duration {
	return e.nextDelay
}

func (e *Exponential) randOffset() time.Duration {
	return time.Millisecond
}

type ExponentialBuilder struct {
}

func (e ExponentialBuilder) Build() Exponential {
	return Exponential{
		nextDelay: 2,
	}
}

func NewExponentialBuilder() ExponentialBuilder {
	return ExponentialBuilder{}
}
