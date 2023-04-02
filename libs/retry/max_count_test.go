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
		name            string
		errCodes        [](*errs.ErrorCode)
		durations       []time.Duration
		maxCount        int
		sleepAwakeCount int
		expectRetries   int
		expectErr       *errs.Error
	}{
		{
			name:     "Stop retry with ClientInteraction err category",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &clientInteractionAlreadyExistsErr},
			durations: []time.Duration{
				shortDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
			},
			maxCount:        10,
			expectRetries:   3,
			expectErr:       errs.NewError(errs.InvalidArgument, ""),
			sleepAwakeCount: 2,
		},
		{
			name:     "Succeed before reaching max count",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, nil},
			durations: []time.Duration{
				shortDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
			},
			maxCount:        10,
			expectRetries:   3,
			expectErr:       nil,
			sleepAwakeCount: 2,
		},
		{
			name:     "Not Succeed, max count retries reached",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &transientTimeoutErr, &outageUnimplementedErr, &outageUnimplementedErr},
			durations: []time.Duration{
				shortDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
				shortDelay*4 + randomOffset,
				longDelay*4 + randomOffset,
				longDelay*8 + randomOffset,
			},
			maxCount:        5,
			expectRetries:   5,
			expectErr:       errs.NewError(outageUnimplementedErr, ""),
			sleepAwakeCount: 5,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			randGen := randgen_test.NewBuiltinRanGen([]int{1})
			shortBackOff := backoff.NewExponentialBuilder().
				MinDelay(shortDelay).
				RandGenerator(randGen).
				Build()
			longBackOff := backoff.NewExponentialBuilder().
				RandGenerator(randGen).
				MinDelay(longDelay).
				ResetOnSuccess(true).
				Build()
			beforeThreadSleepChan := make(chan bool)
			var currDuration time.Duration
			runtime := runtime_test.NewTestRuntime(func(d time.Duration) {
				currDuration = d
				beforeThreadSleepChan <- true
			})

			count := 0
			execute := func() *errs.Error {
				prevCount := count
				count++
				if prevCount < testCase.expectRetries {
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

				assert.Equal(t, testCase.expectRetries, retries)
				assert.Equal(t, testCase.expectErr, err)
			}()

			retry := 1
			for retry <= testCase.sleepAwakeCount {
				<-beforeThreadSleepChan
				assert.Equal(t, count, retry)
				assert.Equal(t, currDuration, testCase.durations[retry-1])
				runtime.Awake()
				retry++
			}
		})
	}
}
