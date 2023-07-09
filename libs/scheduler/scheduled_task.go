package scheduler

import (
	"context"
)

type Task interface {
	execute(ct context.Context) error
}

type ScheduledTask struct {
	id         uint64
	ct         context.Context // cancel context
	scheduler  *Scheduler
	schedule   Schedule
	task       Task
	cancelFunc func()
}

func (s *ScheduledTask) Remove() {
	s.scheduler.removeScheduledTask(s)
}

func (s *ScheduledTask) RunTask() error {
	return s.task.execute(s.ct)
}

func (s *ScheduledTask) Cancel() {
	s.cancelFunc()
}

func (s *ScheduledTask) Schedule() Schedule {
	return s.schedule
}

func NewScheduledTask(
	id uint64,
	ct context.Context,
	scheduler *Scheduler,
	schedule Schedule,
	task Task,
) ScheduledTask {
	ct, cancelFunc := context.WithCancel(ct)
	return ScheduledTask{
		id:         id,
		ct:         ct,
		scheduler:  scheduler,
		schedule:   schedule,
		task:       task,
		cancelFunc: cancelFunc,
	}
}
