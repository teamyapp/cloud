package scheduler

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
)

var logger = telemetry.NewLogger(
	telemetry.NewOrderedColumnLineFormatter([]string{}),
	os.Stdout,
	telemetry.Off,
	[]telemetry.LogInterceptor{},
)

type testTask struct {
	id        uint64
	counter   int
	counterMu sync.Mutex
}

func (t *testTask) Execute(ct context.Context) error {
	t.counterMu.Lock()
	defer t.counterMu.Unlock()
	t.counter++
	return nil
}

func (t *testTask) GetCounter() int {
	t.counterMu.Lock()
	defer t.counterMu.Unlock()
	return t.counter
}

func (t *testTask) GetID() uint64 {
	return t.id
}

func newTestTask(id uint64) *testTask {
	return &testTask{
		counter: 0,
		id:      id,
	}
}

func TestSchedulerSync(t *testing.T) {
	testCase := struct {
		name           string
		scheduledTasks []struct {
			delays []time.Duration
			task   *testTask
		}
		ranTaskIDs []uint64
	}{
		name: "sync ensure task order",
		scheduledTasks: []struct {
			delays []time.Duration
			task   *testTask
		}{
			{
				delays: []time.Duration{30 * time.Millisecond, 30 * time.Millisecond},
				task:   newTestTask(1),
			},
			{
				delays: []time.Duration{10 * time.Millisecond, 10 * time.Millisecond},
				task:   newTestTask(2),
			},
			{
				delays: []time.Duration{1 * time.Millisecond, 1 * time.Millisecond},
				task:   newTestTask(3),
			},
		},

		ranTaskIDs: []uint64{
			3, 3, 2, 2, 1, 1,
		},
	}

	clock := runtime.NewBuiltinClock()
	t.Run(testCase.name, func(t *testing.T) {
		t.Parallel()

		scheduler, err := NewScheduler(logger, clock)
		assert.Nil(t, err)

		subscription := scheduler.SubscribeTaskStart()
		for _, st := range testCase.scheduledTasks {
			ct := context.Background()
			schedule := NewFixedDelaysSchedule(st.delays, clock)
			scheduler.ScheduleTask(ct, schedule, st.task)
		}

		scheduler.Start()
		subscriptionIndex := 0
		for runningTask := range subscription.Output() {
			assert.Equal(t, testCase.ranTaskIDs[subscriptionIndex], runningTask.GetID())
			subscriptionIndex++
			if subscriptionIndex == len(testCase.ranTaskIDs) {
				break
			}
		}
	})

}

func TestSchedulerConcurrent(t *testing.T) {
	// wait for a bit longer, start scheduler before scheduling tasks, try to schedule tasks concurrently
	testCase := struct {
		name           string
		scheduledTasks []struct {
			delays []time.Duration
			task   *testTask
		}
		ranTaskIDs []uint64
	}{
		name: "ensure task order",
		scheduledTasks: []struct {
			delays []time.Duration
			task   *testTask
		}{
			{
				delays: []time.Duration{50 * time.Millisecond, 50 * time.Millisecond},
				task:   newTestTask(1),
			},
			{
				delays: []time.Duration{30 * time.Millisecond, 30 * time.Millisecond},
				task:   newTestTask(2),
			},
			{
				delays: []time.Duration{10 * time.Millisecond, 10 * time.Millisecond},
				task:   newTestTask(3),
			},
		},
		ranTaskIDs: []uint64{
			3, 3, 2, 1, 2, 1,
		},
	}

	clock := runtime.NewBuiltinClock()

	t.Run(testCase.name, func(t *testing.T) {
		t.Parallel()
		scheduler, err := NewScheduler(logger, clock)
		assert.Nil(t, err)

		endSubscription := scheduler.SubscribeTaskFinish()
		scheduler.Start()
		for _, st := range testCase.scheduledTasks {
			ct := context.Background()
			schedule := NewFixedDelaysSchedule(st.delays, clock)
			go func(task *testTask) {
				scheduler.ScheduleTask(ct, schedule, task)
			}(st.task)
		}

		subscriptionIndex := 0
		for output := range endSubscription.Output() {
			assert.Equal(t, testCase.ranTaskIDs[subscriptionIndex], output.GetID())
			subscriptionIndex++
			if subscriptionIndex == len(testCase.ranTaskIDs) {
				break
			}
		}
	})
}
