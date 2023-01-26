package runtime

import "time"

type BuiltInClock struct {
}

func (b *BuiltInClock) Now() time.Time {
	return time.Now()
}

func NewBuiltinClock() BuiltInClock {
	return BuiltInClock{}
}
