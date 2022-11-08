package backoff

import (
	"time"
)

type Instant struct {
}

var _ BackOff = (*Instant)(nil)

func (i *Instant) OnSuccess() {}

func (i *Instant) OnFailure() {}

func (i *Instant) Delay() time.Duration {
	return 0
}

type InstantBuilder struct {
}

func (i InstantBuilder) Build() Instant {
	return Instant{}
}

func NewInstantBuilder() InstantBuilder {
	return InstantBuilder{}
}
