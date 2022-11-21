package runtime_test

import "time"

type TestRuntime struct {
	wakeupChan        chan bool
	beforeThreadSleep func(time.Duration)
}

func (b BuiltInRuntime) Sleep(duration time.Duration) {
	if b.beforeThreadSleep != nil {
		b.beforeThreadSleep(duration)
	}

	<-b.wakeupChan
}

func (b BuiltInRuntime) Awake() {
	b.wakeupChan <- true
}

func NewBuiltInRuntime(beforeThreadSleep func(time.Duration)) *BuiltInRuntime {
	runtime := &BuiltInRuntime{
		beforeThreadSleep: beforeThreadSleep,
		wakeupChan:        make(chan bool),
	}

	return runtime
}
