package runtime_test

import (
	"sync"
	"time"

	"github.com/teamyapp/cloud/libs/runtime"
)

type TestClock struct {
	now   time.Time
	nowMu sync.RWMutex
}

var _ runtime.Clock = (*TestClock)(nil)

func (t *TestClock) Now() time.Time {
	t.nowMu.RLock()
	defer t.nowMu.RUnlock()
	return t.now
}

func (t *TestClock) SetNow(now time.Time) {
	t.nowMu.Lock()
	defer t.nowMu.Unlock()
	t.now = now
}

func (t *TestClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func NewTestClock(now time.Time) *TestClock {
	return &TestClock{
		now: now,
	}
}
