package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/retry/backoff/backoff_test"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestInfinite(t *testing.T) {
	testCases := []struct {
		name    string
		count   int
		retries int
	}{
		{
			name:    "Should retry until succeed",
			count:   0,
			retries: 10,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
t.Parallel()
			beforeThreadSleepChan := make(chan bool)
			backoff := backoff_test.NewExponentialBuilder().Build()
			runtime := runtime_test.NewBuiltInRuntime(func(d time.Duration) {
				beforeThreadSleepChan <- true
			})

			execute := func() error {
				prevCount := testCase.count
				testCase.count++
				if prevCount < testCase.retries {
					return errors.New("some error")
				}
				
				return nil
			}
			infiniteExecutor := NewInfinite(runtime, &backoff)

			go func() {
				retries, err := infiniteExecutor.WithRetry(execute)

				assert.Equal(t, retries, testCase.retries)
				assert.Equal(t, err, nil)
			}()

			retry := 1
			for retry <= testCase.retries {
				<-beforeThreadSleepChan
				assert.Equal(t, testCase.count, retry)
				runtime.Awake()
				retry++
			}
		})
	}
}
