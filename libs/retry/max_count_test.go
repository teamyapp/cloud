package retry

import (
	"errors"
	"time"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/retry/backoff/backoff_test"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestMaxCount(t *testing.T) {

	testCases := []*maxCountTestCase{
		makeMaxCountTestCase("should succeed before reaching max count", 3, nil),
		makeMaxCountTestCase("should not call more than max count retries", 2, errors.New("some error")),
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			go func() {
				retries, err := testCase.maxCountExecutor.WithRetry(testCase.execute)

				assert.Equal(t, retries, 2)
				assert.Equal(t, err, testCase.err)
			}()

			<-testCase.beforeThreadSleepChan
			assert.Equal(t, testCase.count, 1)
			testCase.runtime.Awake()

			<-testCase.beforeThreadSleepChan
			assert.Equal(t, testCase.count, 2)
			testCase.runtime.Awake()
		})
	}
}

type maxCountTestCase struct {
	name                  string
	execute               func() error
	count                 int
	beforeThreadSleepChan chan bool
	maxCountExecutor      MaxCount
	runtime               *runtime_test.Runtime
	err                   error
}

func makeMaxCountTestCase(name string, maxCount int, err error) *maxCountTestCase {
	testCase := maxCountTestCase{
		name:                  name,
		beforeThreadSleepChan: make(chan bool),
		count:                 0,
		err:                   err,
	}

	backoff := backoff_test.NewExponentialBuilder().Build()

	testCase.runtime = runtime_test.NewRuntime(func(d time.Duration) {
		testCase.beforeThreadSleepChan <- true
	})

	testCase.execute = func() error {
		prevCount := testCase.count
		testCase.count++
		if prevCount < 2 {
			return errors.New("some error")
		}
		return nil
	}

	testCase.maxCountExecutor = NewMaxCount(&backoff, maxCount, testCase.runtime)

	return &testCase

}
