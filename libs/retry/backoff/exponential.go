package backoff

import (
	"math"
	"time"

	"github.com/teamyapp/cloud/libs/randgen"
)

type Exponential struct {
	minDelay         time.Duration
	maxDelay         time.Duration
	scalingFactor    int
	randomOffset     int
	randomOffsetUnit time.Duration
	nextDelay        time.Duration
	resetOnSuccess   bool
	randGenerator    randgen.RandomNumberGenerator
}

var _ BackOff = (*Exponential)(nil)

func (e *Exponential) OnSuccess() {
	if e.resetOnSuccess {
		e.nextDelay = e.minDelay
		return
	}

	scaled := time.Duration(math.Pow(float64(e.nextDelay.Milliseconds()), float64(1)/float64(e.scalingFactor)) * float64(time.Millisecond))
	e.nextDelay = max(scaled, e.minDelay) + e.randOffset()
}

func (e *Exponential) OnFailure() {
	scaled := time.Duration(math.Pow(float64(e.nextDelay.Milliseconds()), float64(e.scalingFactor)) * float64(time.Millisecond))
	e.nextDelay = min(scaled, e.maxDelay) + e.randOffset()
}

func (e *Exponential) Delay() time.Duration {
	return e.nextDelay
}

func (e *Exponential) randOffset() time.Duration {
	return time.Duration(e.randGenerator.Intn(e.randomOffset)) * e.randomOffsetUnit
}

type ExponentialBuilder struct {
	minDelay         time.Duration
	maxDelay         time.Duration
	scalingFactor    int
	randomOffset     int
	randomOffsetUnit time.Duration
	resetOnSuccess   bool
	randGenerator    randgen.RandomNumberGenerator
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

func (e ExponentialBuilder) Build() Exponential {
	return Exponential{
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
		minDelay:         200 * time.Millisecond,
		maxDelay:         5 * time.Minute,
		scalingFactor:    2,
		randomOffset:     100,
		randomOffsetUnit: time.Millisecond,
		resetOnSuccess:   false,
		randGenerator:    randgen.BuiltinRanGen{},
	}
}

func min[Num int | uint64 | float32 | float64 | time.Duration](num1 Num, num2 Num) Num {
	if num1 < num2 {
		return num1
	} else {
		return num2
	}
}

func max[Num int | uint64 | float32 | float64 | time.Duration](num1 Num, num2 Num) Num {
	if num1 > num2 {
		return num1
	} else {
		return num2
	}
}
