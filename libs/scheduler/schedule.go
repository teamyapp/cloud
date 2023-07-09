package scheduler

import (
	"time"

	"github.com/teamyapp/cloud/libs/runtime"
)

type Schedule interface {
	nextTimeToRun() time.Time
}

type DelaySchedule struct {
	delay time.Duration
	clock runtime.Clock
}

func (f DelaySchedule) nextTimeToRun() time.Time {
	return f.clock.Now().Add(f.delay)
}

func NewDelaySchedule(delay time.Duration, clock runtime.Clock) DelaySchedule {
	return DelaySchedule{
		delay: delay,
		clock: clock,
	}
}
