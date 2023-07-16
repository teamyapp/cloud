package runtime

import "time"

type BuiltInClock struct {
}

func (b *BuiltInClock) Now() time.Time {
	return time.Now()
}

func (b *BuiltInClock) After(d time.Duration, id uint64) <-chan time.Time {
	return time.After(d)
}

func NewBuiltinClock() *BuiltInClock {
	return &BuiltInClock{}
}
