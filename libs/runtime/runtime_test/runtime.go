package runtime_test

import "time"

type TestRuntime struct {
	wakeupChan        chan bool
	beforeThreadSleep func(time.Duration)
}

func (b TestRuntime) Sleep(duration time.Duration) {
	if b.beforeThreadSleep != nil {
		b.beforeThreadSleep(duration)
	}

	<-b.wakeupChan
}

func (b TestRuntime) Awake() {
	b.wakeupChan <- true
}

func NewTestRuntime(beforeThreadSleep func(time.Duration)) *TestRuntime {
	runtime := &TestRuntime{
		beforeThreadSleep: beforeThreadSleep,
		wakeupChan:        make(chan bool),
	}

	return runtime
}
