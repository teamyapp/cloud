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

func NewInstant() *Instant {
	return &Instant{}
}
