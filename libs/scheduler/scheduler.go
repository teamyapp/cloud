package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/teamyapp/cloud/libs/algo"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/stream"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Scheduler struct {
	logger             telemetry.Logger
	clock              runtime.Clock
	scheduledTasks     *algo.PriorityQueue[ScheduledTask]
	scheduleTaskCh     chan bool
	stopCh             chan bool
	runningTasks       map[uint64]*ScheduledTask
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
	s.logger.InfoWithContext(ct, fmt.Sprintf("enter schedule task: taskID=%d", task.GetID()))
	s.scheduleTaskMu.Lock()
	scheduledTask := NewScheduledTask(ct, s, schedule, task)
	schedule.UpdateNextTimeToRun()
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
			s.scheduleTaskMu.Lock()
			if s.scheduledTasks.Size() == 0 {
				s.scheduleTaskMu.Unlock()
				select {
				case <-s.scheduleTaskCh:
				case <-s.stopCh:
					return
				}
				s.scheduleTaskMu.Lock()
				if s.scheduledTasks.Size() == 0 {
					s.scheduleTaskMu.Unlock()
					continue
				}
			}

			scheduledTask, err := s.scheduledTasks.Peek()
			if err != nil {
				s.scheduleTaskMu.Unlock()
				return
			}

			nextTimeToRun := scheduledTask.Schedule().GetNextTimeToRun()
			now := s.clock.Now()

			if nextTimeToRun.After(now) {
				waitTime := nextTimeToRun.Sub(now)
				s.scheduleTaskMu.Unlock()
				select {
				case <-s.clock.After(waitTime):
				case <-s.scheduleTaskCh:
					continue
				case <-s.stopCh:
					return
				}

				s.scheduleTaskMu.Lock()
				latestScheduledTask, err := s.scheduledTasks.Peek()
				if err != nil {
					s.scheduleTaskMu.Unlock()
					return
				}

				if scheduledTask.task.GetID() != latestScheduledTask.task.GetID() {
					s.logger.Info(fmt.Sprintf("new task is scheduled before current task. Run current task at later time: newTaskID=%v, currentTaskID=%v", latestScheduledTask.task.GetID(), scheduledTask.task.GetID()))
					s.scheduleTaskMu.Unlock()
					continue
				}
			}

			scheduledTask, err = s.scheduledTasks.Pop()
			if err != nil {
				s.scheduleTaskMu.Unlock()
				return
			}

			if scheduledTask.schedule.HasFutureRun() {
				scheduledTask.schedule.UpdateNextTimeToRun()
				s.scheduledTasks.Insert(scheduledTask)
			}

			taskID := scheduledTask.task.GetID()
			s.runningTasks[taskID] = &scheduledTask
			s.scheduleTaskMu.Unlock()

			s.logger.Info(fmt.Sprintf("task %d is starting to run", taskID))
			s.onTaskStartCh <- scheduledTask.task

			go func() {
				scheduledTask.RunTask()
				s.logger.Info(fmt.Sprintf("task %d is finished", taskID))
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
	delete(s.runningTasks, scheduledTask.task.GetID())
	s.scheduledTasks.Remove(scheduledTask)
	s.scheduleTaskMu.Unlock()
	scheduledTask.Cancel()
}

func NewScheduler(
	logger telemetry.Logger,
	clock runtime.Clock,
) (*Scheduler, error) {
	compare := func(first ScheduledTask, second ScheduledTask) algo.Comparison {
		if first.Schedule().GetNextTimeToRun().Before(second.Schedule().GetNextTimeToRun()) {
			return algo.SmallerThan
		}

		if first.Schedule().GetNextTimeToRun().After(second.Schedule().GetNextTimeToRun()) {
			return algo.GreaterThan
		}

		return algo.Equal
	}
	scheduledTasks := algo.NewPriorityQueue[ScheduledTask](compare, nil)
	onTaskStartCh := make(chan Task)
	onTaskFinishCh := make(chan Task)
	return &Scheduler{
		logger:             logger,
		clock:              clock,
		scheduledTasks:     scheduledTasks,
		scheduleTaskCh:     make(chan bool),
		stopCh:             make(chan bool),
		runningTasks:       make(map[uint64]*ScheduledTask),
		scheduleTaskMu:     sync.RWMutex{},
		onTaskStartCh:      onTaskStartCh,
		onTaskFinishCh:     onTaskFinishCh,
		onTaskStartPubSub:  stream.NewPubSub(onTaskStartCh),
		onTaskFinishPubSub: stream.NewPubSub(onTaskFinishCh),
	}, nil
}
