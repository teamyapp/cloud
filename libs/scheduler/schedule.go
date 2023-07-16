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

type FixedDelaysSchedule struct {
	clock          runtime.Clock
	delays         []time.Duration
	nextTimeToRun  time.Time
	nextDelayIndex int
}

func (f *FixedDelaysSchedule) getNextTimeToRun() time.Time {
	return f.nextTimeToRun
}

func (f *FixedDelaysSchedule) updateNextTimeToRun() {
	f.nextTimeToRun = f.clock.Now().Add(f.delays[f.nextDelayIndex])
	f.nextDelayIndex++
}

func (f *FixedDelaysSchedule) HasFutureRun() bool {
	return f.nextDelayIndex < len(f.delays)-1
}

func NewFixedDelaysSchedule(delays []time.Duration, clock runtime.Clock) *FixedDelaysSchedule {
	return &FixedDelaysSchedule{
		clock:          clock,
		delays:         delays,
		nextDelayIndex: 0,
	}
}
