package backoff

import (
	"time"
)

type Uniform struct {
	Instant
	delay time.Duration
}

var _ BackOff = (*Uniform)(nil)

func (u *Uniform) Delay() time.Duration {
	return u.delay
}

type UniformBuilder struct {
	delay time.Duration
}

func (u *UniformBuilder) Delay(delay time.Duration) *UniformBuilder {
	u.delay = delay
	return u
}

func (u *UniformBuilder) Build() *Uniform {
	return &Uniform{
		delay: u.delay,
	}
}

func NewUniformBuilder() *UniformBuilder {
	return &UniformBuilder{
		delay: 200 * time.Millisecond,
	}
}
