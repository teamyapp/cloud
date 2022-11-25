package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/retry/backoff/backoff_test"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestTimeout(t *testing.T) {

	testCases := []struct {
		name    string
		timeout time.Duration
		err     error
		awaits  int
	}{

		{
			name:    "Should succeed before reaching timeout",
			timeout: 17,
			err:     nil,
			awaits:  2,
		},
		{
			name:    "Should get error when reaching timeout",
			timeout: 16,
			err:     errors.New("some error"),
			awaits:  1,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			currentTime := time.Now()
			testClock := runtime_test.NewTestClock(currentTime)
			beforeThreadSleepChan := make(chan bool)
			backoff := backoff_test.NewExponentialBuilder().Build()
			builtinRuntime := runtime_test.NewTestRuntime(func(duration time.Duration) {
				testClock.SetTime(testClock.Now().Add(duration))
				beforeThreadSleepChan <- true
			})
			count := 0
			execute := func() error {
				prevCount := count
				count++
				if prevCount < 2 {
					return errors.New("some error")
				}

				return nil
			}
			timeoutExecutor := NewTimeout(builtinRuntime, &backoff, &testClock, testCase.timeout, func() {
				testClock.SetTime(testClock.Now().Add(2))
			})

			go func() {
				retries, err := timeoutExecutor.WithRetry(execute)

				assert.Equal(t, retries, 2)
				assert.Equal(t, err, testCase.err)
			}()

			for await := 0; await < testCase.awaits; await++ {
				<-beforeThreadSleepChan
				assert.Equal(t, count, await+1)
				builtinRuntime.Awake()
			}
		})
	}
}
