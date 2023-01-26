package backoff

import (
	"time"

	"github.com/teamyapp/cloud/libs/num"
	"github.com/teamyapp/cloud/libs/randgen"
)

type Exponential struct {
	randGenerator    randgen.RandomNumberGenerator
	minDelay         time.Duration
	maxDelay         time.Duration
	scalingFactor    int
	randomOffset     int
	randomOffsetUnit time.Duration
	nextDelay        time.Duration
	resetOnSuccess   bool
}

var _ BackOff = (*Exponential)(nil)

func (e *Exponential) OnSuccess() {
	if e.resetOnSuccess {
		e.nextDelay = e.minDelay
		return
	}

	scaled := e.nextDelay / time.Duration(e.scalingFactor)
	e.nextDelay = num.Max(scaled, e.minDelay)
}

func (e *Exponential) OnFailure() {
	scaled := e.nextDelay * time.Duration(e.scalingFactor)
	e.nextDelay = num.Min(scaled, e.maxDelay)
}

func (e *Exponential) Delay() time.Duration {
	return e.nextDelay + e.randOffset()
}

func (e *Exponential) randOffset() time.Duration {
	return time.Duration(e.randGenerator.RandomInt(e.randomOffset)) * e.randomOffsetUnit
}

type ExponentialBuilder struct {
	randGenerator    randgen.RandomNumberGenerator
	minDelay         time.Duration
	maxDelay         time.Duration
	scalingFactor    int
	randomOffset     int
	randomOffsetUnit time.Duration
	resetOnSuccess   bool
}

func (e ExponentialBuilder) MinDelay(minDelay time.Duration) ExponentialBuilder {
	e.minDelay = minDelay
	return e
}

func (e ExponentialBuilder) MaxDelay(maxDelay time.Duration) ExponentialBuilder {
	e.maxDelay = maxDelay
	return e
}

func (e ExponentialBuilder) ScalingFactor(scalingFactor int) ExponentialBuilder {
	e.scalingFactor = scalingFactor
	return e
}

func (e ExponentialBuilder) RandomOffset(randomOffset int) ExponentialBuilder {
	e.randomOffset = randomOffset
	return e
}

func (e ExponentialBuilder) RandomOffsetUnit(randomOffsetUnit time.Duration) ExponentialBuilder {
	e.randomOffsetUnit = randomOffsetUnit
	return e
}

func (e ExponentialBuilder) ResetOnSuccess(resetOnSuccess bool) ExponentialBuilder {
	e.resetOnSuccess = resetOnSuccess
	return e
}

func (e ExponentialBuilder) RandGenerator(randGenerator randgen.RandomNumberGenerator) ExponentialBuilder {
	e.randGenerator = randGenerator
	return e
}

func (e ExponentialBuilder) Build() *Exponential {
	return &Exponential{
		minDelay:         e.minDelay,
		maxDelay:         e.maxDelay,
		scalingFactor:    e.scalingFactor,
		randomOffset:     e.randomOffset,
		randomOffsetUnit: e.randomOffsetUnit,
		nextDelay:        e.minDelay,
		resetOnSuccess:   e.resetOnSuccess,
		randGenerator:    e.randGenerator,
	}
}

func NewExponentialBuilder() ExponentialBuilder {
	return ExponentialBuilder{
		randGenerator:    randgen.NewBuiltinRanGen(),
		minDelay:         200 * time.Millisecond,
		maxDelay:         5 * time.Minute,
		scalingFactor:    2,
		randomOffset:     100,
		randomOffsetUnit: time.Millisecond,
		resetOnSuccess:   false,
	}
}
