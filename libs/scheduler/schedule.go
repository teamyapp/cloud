package scheduler

import (
	"time"

	"github.com/teamyapp/cloud/libs/runtime"
)

type Schedule interface {
	getNextTimeToRun() time.Time
	updateNextTimeToRun()
	HasFutureRun() bool
}

type OneTimeDelaySchedule struct {
	clock                runtime.Clock
	delay                time.Duration
	nextTimeToRun        time.Time
	scheduleHasFutureRun bool
}

func (f *OneTimeDelaySchedule) getNextTimeToRun() time.Time {
	return f.nextTimeToRun
}

func (f *OneTimeDelaySchedule) updateNextTimeToRun() {
	f.nextTimeToRun = f.clock.Now().Add(f.delay)
	f.scheduleHasFutureRun = true
}

func (f *OneTimeDelaySchedule) HasFutureRun() bool {
	return f.scheduleHasFutureRun
}

func NewOneTimeDelaySchedule(delay time.Duration, clock runtime.Clock) *OneTimeDelaySchedule {
	return &OneTimeDelaySchedule{
		clock: clock,
		delay: delay,
	}
}
