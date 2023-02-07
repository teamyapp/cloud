package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestMaxCount(t *testing.T) {
	testCases := []struct {
		name     string
		maxCount int
		err      *errs.Error
	}{
		{
			name:     "Succeed before reaching max count",
			maxCount: 3,
			err:      nil,
		},
		{
			name:     "Not exceed max count retries",
			maxCount: 2,
			err:      &errs.Error{Code: errs.Serialization},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			backoff := backoff.NewExponentialBuilder().Build()
			beforeThreadSleepChan := make(chan bool)
			builtinRuntime := runtime_test.NewTestRuntime(func(d time.Duration) {
				beforeThreadSleepChan <- true
			})
			count := 0
			execute := func() *errs.Error {
				prevCount := count
				count++
				if prevCount < 2 {
					return &errs.Error{
						Code: errs.Serialization,
					}
				}

				return nil
			}

			maxCountExecutor := NewMaxCount(builtinRuntime, backoff, testCase.maxCount)

			go func() {
				retries, err := maxCountExecutor.WithRetry(execute)

				assert.Equal(t, 2, retries)
				assert.Equal(t, testCase.err, err)
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
