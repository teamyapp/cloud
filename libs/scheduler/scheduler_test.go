package scheduler

import (
	"context"
	"fmt"
	"os"
	"sync"
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
					task:  newTestTask(1),
				},
				{
					delay: 3 * time.Second,
					task:  newTestTask(2),
				},
				{
					delay: 1 * time.Second,
					task:  newTestTask(3),
				},
			},
		},
	}

	currentTime := time.Now()
	testClock := runtime_test.NewTestClock(currentTime)
	fmt.Printf("Start a new test suit\n\n\n\n")

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			scheduler, err := NewScheduler(testClock)
			assert.Nil(t, err)

			scheduler.Start()
			subscription := scheduler.SubscribeTaskFinish()
			wg := &sync.WaitGroup{}
			for _, st := range testCase.scheduledTasks {
				ct := context.Background()
				schedule := NewOneTimeDelaySchedule(st.delay, testClock)
				wg.Add(1)
				go func(task *testTask) {
					defer wg.Done()
					scheduler.ScheduleTask(ct, schedule, task)
				}(st.task)
			}
			wg.Wait()
			testClock.SetNow(currentTime.Add(1 * time.Second))

			// runningTaskId := <-scheduler.OnTaskStart()
			// assert.Equal(t, uint64(1), runningTaskId)
			runningTask := <-subscription.Output()
			assert.Equal(t, uint64(3), runningTask.getID())
			assert.Equal(t, 1, testCase.scheduledTasks[2].task.counter)
		})
	}
}
