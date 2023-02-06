package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestInfinite(t *testing.T) {
	testCases := []struct {
		name    string
		retries int
	}{
		{
			name:    "Should retry until succeed",
			retries: 10,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			beforeThreadSleepChan := make(chan bool)
			backoff := backoff.NewExponentialBuilder().Build()
			runtime := runtime_test.NewTestRuntime(func(d time.Duration) {
				beforeThreadSleepChan <- true
			})

			count := 0
			execute := func() *errs.Error {
				prevCount := count
				count++
				if prevCount < testCase.retries {
					return &errs.Error{
						Code: errs.Unknown,
					}
				}

				return nil
			}
			infiniteExecutor := NewInfinite(runtime, backoff)

			go func() {
				retries, err := infiniteExecutor.WithRetry(execute)

				assert.Equal(t, retries, testCase.retries)
				assert.Equal(t, err, nil)
			}()

			retry := 1
			for retry <= testCase.retries {
				<-beforeThreadSleepChan
				assert.Equal(t, count, retry)
				runtime.Awake()
				retry++
			}
		})
	}
}
