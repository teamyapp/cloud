package retry

import (
	"errors"
	"time"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/retry/backoff/backoff_test"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestTimeout(t *testing.T) {

	testCases := []*timeoutTestCase{
		makeTimeoutTestCase("should succeed before reaching timeout", 25*time.Millisecond, nil),
		makeTimeoutTestCase("should get error when reaching timeout", 24*time.Millisecond, errors.New("some error")),
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			go func() {
				retries, err := testCase.timeoutExecutor.WithRetry(testCase.execute)

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

type timeoutTestCase struct {
	name                  string
	execute               func() error
	count                 int
	beforeThreadSleepChan chan bool
	timeoutExecutor       Timeout
	runtime               *runtime_test.Runtime
	err                   error
}

func makeTimeoutTestCase(name string, maxDuration time.Duration, err error) *timeoutTestCase {
	testCase := timeoutTestCase{
		name:                  name,
		beforeThreadSleepChan: make(chan bool),
		count:                 0,
		err:                   err,
	}

	currentTime := time.Now()
	testClock := runtime_test.TestClock{}

	testClock.SetTime(currentTime)

	backoff := backoff_test.NewExponentialBuilder().Build()

	testCase.runtime = runtime_test.NewRuntime(func(duration time.Duration) {
		testClock.SetTime(testClock.Now().Add(duration))
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

	testCase.timeoutExecutor = NewTimeout(&backoff, maxDuration, &testClock, func() {
		testClock.SetTime(testClock.Now().Add(2 * time.Millisecond))
	}, testCase.runtime)

	return &testCase

}
