package runtime_test

import "time"

type Runtime struct {
	wakeupChan     chan bool
	beforeThreadSleep func(time.Duration)
}

func (t Runtime) Sleep(duration time.Duration) {
	if t.beforeThreadSleep != nil {
		(t.beforeThreadSleep)(duration)
	}

	<-t.wakeupChan
}

func (t Runtime) Awake() {
	t.wakeupChannel <- true
}

func NewRuntime(beforeThreadSleep func(time.Duration)) *Runtime {
	runtime := &Runtime{
		beforeThreadSleep: beforeThreadSleep,
		wakeupChan:     make(chan bool),
	}

	return runtime
}
