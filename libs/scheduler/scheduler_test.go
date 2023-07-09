package scheduler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/app/dao/daotest"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/dbtest"
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

	allocatedRanges := []entity.AllocatedRange{
		{
			Key:        "scheduledTaskID",
			RangeEnd:   0,
			NextNumber: 1,
		},
	}

	currentTime := time.Now()
	testClock := runtime_test.NewTestClock(currentTime)

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			inMemoryDB := dbtest.NewInMemoryDB()
			inMemoryDB.InitTable(
				daotest.AllocatedRangeTableName,
				collect.Map(allocatedRanges, func(allocatedRange entity.AllocatedRange, index int) interface{} {
					return allocatedRange
				}))

			mockAllocatedRange := daotest.NewAllocatedRange(inMemoryDB)

			scheduler, err := NewScheduler(gen.NewUniqueNumberFactory(logger, mockAllocatedRange, 0), testClock)
			assert.Nil(t, err)

			scheduler.Start()
			<-time.After(1 * time.Second)
			for _, st := range testCase.scheduledTasks {

				ct := context.Background()

				schedule := NewDelaySchedule(st.delay, testClock)

				scheduler.ScheduleTask(ct, schedule, st.task)
				// <-time.After(500 * time.Millisecond)
			}

			testClock.SetNow(currentTime.Add(1 * time.Second))
			<-time.After(1 * time.Second)

			assert.Equal(t, 1, testCase.scheduledTasks[2].task.counter)
		})
	}
}
