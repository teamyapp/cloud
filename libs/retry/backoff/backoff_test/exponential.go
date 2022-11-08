package backoff_test

import (
	"math"
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
	e.nextDelay = time.Duration(math.Pow(float64(e.nextDelay.Milliseconds()), float64(2)) * float64(time.Millisecond))
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
		nextDelay: 2 * time.Millisecond,
	}
}

func NewExponentialBuilder() ExponentialBuilder {
	return ExponentialBuilder{}
}
