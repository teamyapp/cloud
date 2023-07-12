package scheduler

import (
	"context"
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
}

func (t *testTask) execute(ct context.Context) error {
	t.counter++
	return nil
}

func newTestTask() *testTask {
	return &testTask{
		counter: 0,
	}
}

func TestScheduler(t *testing.T) {
	testCases := []struct {
		name           string
		scheduledTasks []struct {
			delay time.Duration
			task  *testTask
		}
	}{
		{
			name: "test schedulers",
			scheduledTasks: []struct {
				delay time.Duration
				task  *testTask
			}{
				{
					delay: 5 * time.Second,
					task:  newTestTask(),
				},
				{
					delay: 3 * time.Second,
					task:  newTestTask(),
				},
				{
					delay: 1 * time.Second,
					task:  newTestTask(),
				},
			},
		},
	}

	currentTime := time.Now()
	testClock := runtime_test.NewTestClock(currentTime)

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			scheduler, err := NewScheduler(testClock)
			assert.Nil(t, err)

			scheduler.Start()
			subscription := scheduler.SubscribeTaskFinish()
			for _, st := range testCase.scheduledTasks {
				ct := context.Background()

				schedule := NewOneTimeDelaySchedule(st.delay, testClock)

				go func(task *testTask) {
					scheduler.ScheduleTask(ct, schedule, task)
				}(st.task)
			}

			testClock.SetNow(currentTime.Add(1 * time.Second))

			// runningTaskId := <-scheduler.OnTaskStart()
			// assert.Equal(t, uint64(1), runningTaskId)
			runningTaskId := <-subscription.Output()
			assert.Equal(t, uint64(3), runningTaskId)
			assert.Equal(t, 1, testCase.scheduledTasks[2].task.counter)
		})
	}
}
