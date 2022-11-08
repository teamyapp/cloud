package runtime

import "time"

type Clock interface {
	Now() time.Time
}

type BuiltInClock struct {
}

func (b *BuiltInClock) Now() time.Time {
	return time.Now()
}
