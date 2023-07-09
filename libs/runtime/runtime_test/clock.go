package runtime_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/teamyapp/cloud/libs/algo"
)

type After struct {
	deadline time.Time
	afterCh  chan time.Time
}

type TestClock struct {
	now    time.Time
	nowMu  sync.RWMutex
	afters *algo.PriorityQueue[After]
}

func (t *TestClock) Now() time.Time {
	t.nowMu.RLock()
	defer t.nowMu.RUnlock()
	return t.now
}

func (t *TestClock) SetNow(now time.Time) {
	t.nowMu.Lock()
	t.now = now
	afters := []After{}

	for t.afters.Size() > 0 {
		after, err := t.afters.Peek()
		if err != nil {
			break
		}

		if after.deadline.After(t.now) {
			break
		}

		after, err = t.afters.Pop()

		if err != nil {
			break
		}
		afters = append(afters, after)
	}

	t.nowMu.Unlock()

	for _, after := range afters {
		select {
		case after.afterCh <- t.now:
		default:
		}
	}

	fmt.Printf("111111: %v\n", now)
}

func (t *TestClock) After(d time.Duration) <-chan time.Time {
	afterCh := make(chan time.Time)
	after := After{
		deadline: t.now.Add(d),
		afterCh:  afterCh,
	}
	t.nowMu.Lock()
	defer t.nowMu.Unlock()
	t.afters.Insert(after)
	return afterCh
}

func NewTestClock(now time.Time) *TestClock {
	compare := func(a, b After) algo.Comparison {
		if a.deadline.Before(b.deadline) {
			return algo.SmallerThan
		}

		if a.deadline.After(b.deadline) {
			return algo.GreaterThan
		}

		return algo.Equal
	}

	afters := algo.NewPriorityQueue[After](compare, nil)

	return &TestClock{
		now:    now,
		afters: afters,
	}
}
