package scheduler

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
	"github.com/teamyapp/cloud/libs/telemetry"
)

var logger = telemetry.NewLogger(
	telemetry.NewOrderedColumnLineFormatter([]string{}),
	os.Stdout,
	telemetry.Info,
	[]telemetry.LogInterceptor{},
)

type testTask struct {
	counter int
	id      uint64
}

func (t *testTask) execute(ct context.Context) error {
	t.counter++
	return nil
}

func (t *testTask) getID() uint64 {
	return t.id
}

func newTestTask(id uint64) *testTask {
	return &testTask{
		counter: 0,
		id:      id,
	}
}

func TestSchedulerSyncEnsureTaskOrder(t *testing.T) {
	testCases := []struct {
		name           string
		scheduledTasks []struct {
			delays []time.Duration
			task   *testTask
		}
		ranTasks []struct {
			taskID         uint64
			currentCounter int
		}
	}{
		{
			name: "test schedulers sync ensure task order",

			scheduledTasks: []struct {
				delays []time.Duration
				task   *testTask
			}{
				{
					delays: []time.Duration{5},
					task:   newTestTask(1),
				},
				{
					delays: []time.Duration{3},
					task:   newTestTask(2),
				},
				{
					delays: []time.Duration{1},
					task:   newTestTask(3),
				},
			},

			ranTasks: []struct {
				taskID         uint64
				currentCounter int
			}{
				{
					taskID:         3,
					currentCounter: 1,
				},
				{
					taskID:         2,
					currentCounter: 1,
				},
				{
					taskID:         1,
					currentCounter: 1,
				},
			},
		},
	}

	fmt.Printf("Start a new test suit==================\n")
	clock := runtime_test.NewTestClock(time.Now())
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			scheduler, err := NewScheduler(clock)
			assert.Nil(t, err)

			subscription := scheduler.SubscribeTaskStart()
			for _, st := range testCase.scheduledTasks {
				ct := context.Background()
				schedule := NewFixedDelaysSchedule(st.delays, clock)
				scheduler.ScheduleTask(ct, schedule, st.task)
			}

			clock.SetNow(time.Now().Add(1 * time.Millisecond))
			scheduler.Start()
			subscriptionIndex := 0
			for runningTask := range subscription.Output() {
				assert.Equal(t, testCase.ranTasks[subscriptionIndex].taskID, runningTask.getID())
				subscriptionIndex++
				if subscriptionIndex == len(testCase.ranTasks) {
					break
				}
			}
		})
	}
}

func TestSchedulerConcurrentlyScheduled(t *testing.T) {

}

func TestSchedulerConcurrentEnsureTaskOrder(t *testing.T) {

}
