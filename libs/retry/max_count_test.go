package retry

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen/randgen_test"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
	"github.com/teamyapp/cloud/libs/telemetry"
)

func TestMaxCount(t *testing.T) {
	var transientTimeoutErr errs.ErrorCode = errs.Timeout
	var clientInteractionAlreadyExistsErr errs.ErrorCode = errs.AlreadyExists
	var outageUnimplementedErr errs.ErrorCode = errs.Unimplemented

	testCases := []struct {
		name           string
		errCodes       [](*errs.ErrorCode)
		durations      []time.Duration
		maxCount       int
		sleepAwakeLoop int
		expectRetries     int
		expectErr         *errs.Error
	}{
		{
			name:     "Stop retry with ClientInteraction err category",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &clientInteractionAlreadyExistsErr},
			durations: []time.Duration{
				401000000, 501000000, 501000000,
			},
			maxCount:       10,
			resRetries:     3,
			resErr:         &errs.Error{Code: errs.InvalidArgument},
			sleepAwakeLoop: 2,
		},
		{
			name:     "Succeed before reaching max count",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, nil},
			durations: []time.Duration{
				401000000, 
				501000000, 
				501000000,
			},
			maxCount:       10,
			resRetries:     3,
			resErr:         nil,
			sleepAwakeCount: 2,
		},
		{
			name:     "Not Succeed, max count retries reached",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &transientTimeoutErr, &outageUnimplementedErr, &outageUnimplementedErr},
			durations: []time.Duration{
				401000000, 
				501000000, 
				801000000, 
				1001000000, 
				2001000000,
			},
			maxCount:       5,
			resRetries:     5,
			resErr:         &errs.Error{Code: outageUnimplementedErr},
			sleepAwakeLoop: 5,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			randGen := randgen_test.NewBuiltinRanGen([]int{1})
			shortBackOff := backoff.NewExponentialBuilder().RandGenerator(randGen).Build()
			longBackOff := backoff.NewExponentialBuilder().
				RandGenerator(randGen).
				MinDelay(250 * time.Millisecond).
				ResetOnSuccess(true).
				Build()
			beforeThreadSleepChan := make(chan bool)
			var curD time.Duration
			runtime := runtime_test.NewTestRuntime(func(d time.Duration) {
				curD = d
				beforeThreadSleepChan <- true
			})

			count := 0
			execute := func() *errs.Error {
				prevCount := count
				count++
				if prevCount < testCase.resRetries {
					if testCase.errCodes[prevCount] == nil {
						return nil
					}

					return &errs.Error{
						Code: *testCase.errCodes[prevCount],
					}
				}

				return nil
			}

			ct := context.Background()
			lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
			logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
			beforeSkipRetry := func() {
				beforeThreadSleepChan <- true
			}
			maxCountExecutor := NewMaxCount(
				telemetry.NewDataCollector(logger),
				shortBackOff,
				longBackOff,
				runtime,
				testCase.maxCount,
				&beforeSkipRetry,
			)

			go func() {
				retries, err := maxCountExecutor.WithRetry(ct, execute)

				assert.Equal(t, testCase.resRetries, retries)
				assert.Equal(t, testCase.resErr, err)
			}()

			retry := 1
			for retry <= testCase.sleepAwakeLoop {
				<-beforeThreadSleepChan
				assert.Equal(t, count, retry)
				assert.Equal(t, curD, testCase.durations[retry-1])
				runtime.Awake()
				retry++
			}
		})
	}
}
