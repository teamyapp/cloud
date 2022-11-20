package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/retry/backoff/backoff_test"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestInfinite_flow(t *testing.T) {
	testCases := []struct {
		name                  string
		execute               func() error
		count                 int
		beforeThreadSleepChan chan bool
		infiniteExecutor      Infinite
		runtime               *runtime_test.BuiltInRuntime
		retries               int
	}{
		{
			name:                  "Should retry until succeed",
			beforeThreadSleepChan: make(chan bool),
			count:                 0,
			retries:               10,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase := &testCase

			backoff := backoff_test.NewExponentialBuilder().Build()
			testCase.runtime = runtime_test.NewBuiltInRuntime(func(d time.Duration) {
				testCase.beforeThreadSleepChan <- true
			})

			testCase.execute = func() error {
				prevCount := testCase.count
				testCase.count++
				if prevCount < testCase.retries {
					return errors.New("some error")
				}
				return nil
			}
			testCase.infiniteExecutor = NewInfinite(testCase.runtime, &backoff)

			go func() {
				retries, err := testCase.infiniteExecutor.WithRetry(testCase.execute)

				assert.Equal(t, retries, testCase.retries)
				assert.Equal(t, err, nil)
			}()

			retry := 1
			for retry <= testCase.retries {
				<-testCase.beforeThreadSleepChan
				assert.Equal(t, testCase.count, retry)
				testCase.runtime.Awake()
				retry++
			}
		})
	}
}
