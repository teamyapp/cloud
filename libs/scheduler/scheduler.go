package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/teamyapp/cloud/libs/algo"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/stream"
)

type Scheduler struct {
	scheduledTasks     *algo.PriorityQueue[ScheduledTask]
	scheduleTaskCh     chan bool
	stopCh             chan bool
	runningTasks       map[uint64]*ScheduledTask
	clock              runtime.Clock
	scheduleTaskMu     sync.RWMutex
	onTaskStartCh      chan uint64
	onTaskFinishCh     chan uint64
	onTaskStartPubSub  *stream.PubSub[uint64]
	onTaskFinishPubSub *stream.PubSub[uint64]
	nextTaskID         uint64
}

func (s *Scheduler) SubscribeTaskStart() *stream.Subscription[uint64] {
	return s.onTaskStartPubSub.Subscribe()
}

func (s *Scheduler) SubscribeTaskFinish() *stream.Subscription[uint64] {
	return s.onTaskFinishPubSub.Subscribe()
}

func (s *Scheduler) ScheduleTask(
	ct context.Context,
	schedule Schedule,
	task Task,
) (*ScheduledTask, *errs.Error) {
	s.scheduleTaskMu.Lock()
	scheduledTaskID := s.nextTaskID
	s.nextTaskID++
	scheduledTask := NewScheduledTask(scheduledTaskID, ct, s, schedule, task)
	schedule.updateNextTimeToRun()
	s.scheduledTasks.Insert(scheduledTask)
	s.scheduleTaskMu.Unlock()
	s.scheduleTaskCh <- true
	return &scheduledTask, nil
}

func (s *Scheduler) Start() {
	go func() {
		for {
			fmt.Printf("new scheduler loop\n")
			s.scheduleTaskMu.Lock()
			if s.scheduledTasks.Size() == 0 {
				fmt.Printf("no task scheduled\n")
				s.scheduleTaskMu.Unlock()
				select {
				case <-s.scheduleTaskCh:
					fmt.Printf("task scheduled(1)\n")
				case <-s.stopCh:
					fmt.Printf("stopped(1)\n")
					return
				}
				s.scheduleTaskMu.Lock()
				if s.scheduledTasks.Size() == 0 {
					s.scheduleTaskMu.Unlock()
					continue
				}
			}

			fmt.Printf("has task scheduled\n")
			scheduledTask, err := s.scheduledTasks.Peek()
			if err != nil {
				s.scheduleTaskMu.Unlock()
				return
			}

			fmt.Printf("next time to run %v, pq size = %d\n", scheduledTask.Schedule().getNextTimeToRun(), s.scheduledTasks.Size())
			nextTimeToRun := scheduledTask.Schedule().getNextTimeToRun()
			now := s.clock.Now()

			if nextTimeToRun.After(now) {
				fmt.Printf("task will run later: task id=%v, time=%v\n", scheduledTask.id, now)
				waitTime := nextTimeToRun.Sub(now)
				s.scheduleTaskMu.Unlock()
				select {
				case <-s.clock.After(waitTime):
					fmt.Printf("waited %v for task %d\n", waitTime, scheduledTask.id)
				case <-s.scheduleTaskCh:
					fmt.Printf("task scheduled(2)\n")
					continue
				case <-s.stopCh:
					fmt.Printf("stopped(2)\n")
					return
				}

				s.scheduleTaskMu.Lock()
				latestScheduledTask, err := s.scheduledTasks.Peek()
				if err != nil {
					s.scheduleTaskMu.Unlock()
					return
				}

				if scheduledTask.id != latestScheduledTask.id {
					s.scheduleTaskMu.Unlock()
					continue
				}
			} else {
				fmt.Printf("task will run now\n")
			}

			scheduledTask, err = s.scheduledTasks.Pop()
			if err != nil {
				s.scheduleTaskMu.Unlock()
				return
			}

			// if scheduledTask.schedule.HasFutureRun() {
			// 	scheduledTask.schedule.updateNextTimeToRun()
			// 	s.scheduledTasks.Insert(scheduledTask)
			// 	// ?s.scheduleTaskCh <- true
			// }

			s.runningTasks[scheduledTask.id] = &scheduledTask
			s.scheduleTaskMu.Unlock()
			go func() {
				fmt.Printf("running task %d\n", scheduledTask.id)

				s.onTaskStartCh <- scheduledTask.id
				scheduledTask.RunTask()
				fmt.Printf("task %d finished\n", scheduledTask.id)
				s.onTaskFinishCh <- scheduledTask.id

				s.scheduleTaskMu.Lock()
				defer s.scheduleTaskMu.Unlock()
				delete(s.runningTasks, scheduledTask.id)
			}()
		}
	}()
}

func (s *Scheduler) GetRunningTasks() map[uint64]*ScheduledTask {
	s.scheduleTaskMu.RLock()
	defer s.scheduleTaskMu.RUnlock()
	return s.runningTasks
}

func (s *Scheduler) OnTaskStart() <-chan uint64 {
	return s.onTaskStartCh
}

func (s *Scheduler) OnTaskFinish() <-chan uint64 {
	return s.onTaskFinishCh
}

func (s *Scheduler) Stop() {
	s.stopCh <- true

	s.scheduleTaskMu.Lock()
	runningTasks := s.runningTasks
	s.scheduleTaskMu.Unlock()

	for _, scheduledTask := range runningTasks {
		scheduledTask.Cancel()
	}

	s.scheduleTaskMu.Lock()
	defer s.scheduleTaskMu.Unlock()
	s.runningTasks = map[uint64]*ScheduledTask{}
}

func (s *Scheduler) removeScheduledTask(scheduledTask *ScheduledTask) {
	s.scheduleTaskMu.Lock()
	delete(s.runningTasks, scheduledTask.id)
	s.scheduledTasks.Remove(scheduledTask)
	s.scheduleTaskMu.Unlock()
	scheduledTask.Cancel()
}

func NewScheduler(
	clock runtime.Clock,
) (*Scheduler, error) {
	compare := func(a ScheduledTask, b ScheduledTask) algo.Comparison {
		if a.Schedule().getNextTimeToRun().Before(b.Schedule().getNextTimeToRun()) {
			return algo.SmallerThan
		}

		if a.Schedule().getNextTimeToRun().After(b.Schedule().getNextTimeToRun()) {
			return algo.GreaterThan
		}

		return algo.Equal
	}

	scheduledTasks := algo.NewPriorityQueue[ScheduledTask](compare, nil)
	onTaskStartCh := make(chan uint64)
	onTaskFinishCh := make(chan uint64)
	return &Scheduler{
		scheduledTasks:     scheduledTasks,
		scheduleTaskCh:     make(chan bool),
		stopCh:             make(chan bool),
		onTaskStartCh:      onTaskStartCh,
		onTaskFinishCh:     onTaskFinishCh,
		onTaskStartPubSub:  stream.NewPubSub(onTaskStartCh),
		onTaskFinishPubSub: stream.NewPubSub(onTaskFinishCh),
		runningTasks:       make(map[uint64]*ScheduledTask),
		nextTaskID:         1,
		clock:              clock,
	}, nil
}
