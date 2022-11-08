package runtime_test

import "time"

type Runtime struct {
	beforeThreadSleep func(time.Duration)
	wakeupChannel     chan bool
}

func (t Runtime) Sleep(duration time.Duration) {
	if t.beforeThreadSleep != nil {
		(t.beforeThreadSleep)(duration)
	}

	<-t.wakeupChannel
}

func (t Runtime) Awake() {
	t.wakeupChannel <- true
}

func NewRuntime(beforeThreadSleep func(time.Duration)) *Runtime {
	runtime := &Runtime{
		beforeThreadSleep: beforeThreadSleep,
		wakeupChannel:     make(chan bool),
	}

	return runtime
}
