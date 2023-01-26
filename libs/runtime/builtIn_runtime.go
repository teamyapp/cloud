package runtime

import "time"

type BuiltInRuntime struct {
}

func (b BuiltInRuntime) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func NewBuiltInRuntime() BuiltInRuntime {
	return BuiltInRuntime{}
}
