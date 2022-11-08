package runtime

import "time"

type Runtime interface {
	Sleep(duration time.Duration)
}

type BuiltInRuntime struct {
}

func (b BuiltInRuntime) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
