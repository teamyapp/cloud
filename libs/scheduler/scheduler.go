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
	onTaskStartCh      chan Task
	onTaskFinishCh     chan Task
	onTaskStartPubSub  *stream.PubSub[Task]
	onTaskFinishPubSub *stream.PubSub[Task]
}

func (s *Scheduler) SubscribeTaskStart() *stream.Subscription[Task] {
	return s.onTaskStartPubSub.Subscribe()
}

func (s *Scheduler) SubscribeTaskFinish() *stream.Subscription[Task] {
	return s.onTaskFinishPubSub.Subscribe()
}

func (s *Scheduler) ScheduleTask(
	ct context.Context,
	schedule Schedule,
	task Task,
) (*ScheduledTask, *errs.Error) {
	fmt.Printf("schedule a task[%d]++++\n", task.getID())
	s.scheduleTaskMu.Lock()
	scheduledTask := NewScheduledTask(ct, s, schedule, task)
	schedule.updateNextTimeToRun()
	fmt.Printf("task inserted[%d]++++\n", task.getID())
	s.scheduledTasks.Insert(scheduledTask)
	s.scheduleTaskMu.Unlock()

	select {
	case s.scheduleTaskCh <- true:
	default:
	}

	return &scheduledTask, nil
}

func (s *Scheduler) Start() {
	go func() {
		for {
			fmt.Printf("        new scheduler loop\n")
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

			fmt.Printf("next time to run %v, pq size = %d, pop task id %d\n", scheduledTask.Schedule().getNextTimeToRun(), s.scheduledTasks.Size(), scheduledTask.task.getID())
			nextTimeToRun := scheduledTask.Schedule().getNextTimeToRun()
			now := s.clock.Now()

			if nextTimeToRun.After(now) {
				fmt.Printf("task will run later: task id=%v, time=%v\n", scheduledTask.task.getID(), now)
				waitTime := nextTimeToRun.Sub(now)
				s.scheduleTaskMu.Unlock()
				fmt.Printf("set after: %v\n", waitTime)
				select {
				case <-s.clock.After(waitTime, scheduledTask.task.getID()):
					fmt.Printf("waited %v for task %d\n", waitTime, scheduledTask.task.getID())
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

				if scheduledTask.task.getID() != latestScheduledTask.task.getID() {
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

			if scheduledTask.schedule.HasFutureRun() {
				scheduledTask.schedule.updateNextTimeToRun()
				s.scheduledTasks.Insert(scheduledTask)
			}

			taskID := scheduledTask.task.getID()
			s.runningTasks[taskID] = &scheduledTask
			s.scheduleTaskMu.Unlock()

			fmt.Printf("running task %d, started\n", taskID)
			s.onTaskStartCh <- scheduledTask.task

			go func() {
				scheduledTask.RunTask()
				fmt.Printf("task %d finished\n", taskID)
				s.onTaskFinishCh <- scheduledTask.task

				s.scheduleTaskMu.Lock()
				defer s.scheduleTaskMu.Unlock()
				delete(s.runningTasks, taskID)
			}()
		}
	}()
}

func (s *Scheduler) GetRunningTasks() map[uint64]*ScheduledTask {
	s.scheduleTaskMu.RLock()
	defer s.scheduleTaskMu.RUnlock()
	return s.runningTasks
}

func (s *Scheduler) OnTaskStart() <-chan Task {
	return s.onTaskStartCh
}

func (s *Scheduler) OnTaskFinish() <-chan Task {
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
	delete(s.runningTasks, scheduledTask.task.getID())
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
	onTaskStartCh := make(chan Task)
	onTaskFinishCh := make(chan Task)
	return &Scheduler{
		scheduledTasks:     scheduledTasks,
		scheduleTaskCh:     make(chan bool),
		stopCh:             make(chan bool),
		onTaskStartCh:      onTaskStartCh,
		onTaskFinishCh:     onTaskFinishCh,
		onTaskStartPubSub:  stream.NewPubSub(onTaskStartCh),
		onTaskFinishPubSub: stream.NewPubSub(onTaskFinishCh),
		runningTasks:       make(map[uint64]*ScheduledTask),
		clock:              clock,
	}, nil
}
