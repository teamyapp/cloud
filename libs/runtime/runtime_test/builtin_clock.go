package runtime_test

import "time"

type BuiltInClock struct {
	time time.Time
}

func (b *BuiltInClock) Now() time.Time {
	return b.time
}

func (b *BuiltInClock) SetTime(time time.Time) {
	b.time = time
}

func NewBuiltinClock(time time.Time) BuiltInClock {
	return BuiltInClock{
		time: time,
	}
}
