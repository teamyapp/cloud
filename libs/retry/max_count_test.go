package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/retry/backoff/backoff_test"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestMaxCount_flow(t *testing.T) {
	testCases := []struct {
		name     string
		maxCount int
		err      error
	}{
		{
			name:     "Should succeed before reaching max count",
			maxCount: 3,
			err:      nil,
		},
		{
			name:     "Should not call more than max count retries",
			maxCount: 2,
			err:      errors.New("some error"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			backoff := backoff_test.NewExponentialBuilder().Build()
			beforeThreadSleepChan := make(chan bool)
			builtinRuntime := runtime_test.NewBuiltInRuntime(func(d time.Duration) {
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

			maxCountExecutor := NewMaxCount(builtinRuntime, &backoff, testCase.maxCount)

			go func() {
				retries, err := maxCountExecutor.WithRetry(execute)

				assert.Equal(t, retries, 2)
				assert.Equal(t, err, testCase.err)
			}()

			<-beforeThreadSleepChan
			assert.Equal(t, count, 1)
			builtinRuntime.Awake()

			<-beforeThreadSleepChan
			assert.Equal(t, count, 2)
			builtinRuntime.Awake()
		})
	}
}
