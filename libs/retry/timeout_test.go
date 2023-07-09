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

func TestTimeout(t *testing.T) {
	var transientTimeoutErr = errs.Timeout
	var clientInteractionAlreadyExistsErr = errs.AlreadyExists
	var outageUnimplementedErr = errs.Unimplemented
	currentTime := time.Now()
	testClock := runtime_test.NewTestClock(currentTime)
	testCases := []struct {
		name            string
		errCodes        []*errs.ErrorCode
		durations       []time.Duration
		timeout         time.Duration
		executeDuration []time.Duration
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
			timeout:         10 * time.Second,
			executeDuration: []time.Duration{2 * time.Millisecond, 3 * time.Millisecond, 4 * time.Millisecond},
			expectRetries:   3,
			expectErr:       errs.NewError(errs.InvalidArgument, ""),
			sleepAwakeCount: 2,
		},
		{
			name:     "Succeed before reaching max timeout",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, nil},
			durations: []time.Duration{
				shortDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
			},
			timeout:         10 * time.Second,
			executeDuration: []time.Duration{2 * time.Millisecond, 3 * time.Millisecond, 4 * time.Millisecond},
			expectRetries:   3,
			expectErr:       nil,
			sleepAwakeCount: 2,
		},
		{
			name:     "Not Succeed, max timeout reached",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &transientTimeoutErr},
			durations: []time.Duration{
				shortDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
				longDelay*2 + randomOffset,
			},
			timeout:         10 * time.Second,
			executeDuration: []time.Duration{2 * time.Second, 3 * time.Second, 4 * time.Second},
			expectRetries:   3,
			expectErr:       errs.NewError(transientTimeoutErr, ""),
			sleepAwakeCount: 2,
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
				MinDelay(250 * time.Millisecond).
				ResetOnSuccess(true).
				Build()
			beforeThreadSleepChan := make(chan bool)
			var currDuration time.Duration
			runtime := runtime_test.NewTestRuntime(func(d time.Duration) {
				currDuration = d
				testClock.SetNow(testClock.Now().Add(d))
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

					return errs.NewError(*testCase.errCodes[prevCount], "")
				}

				return nil
			}

			ct := context.Background()
			lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
			logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
			beforeSkipRetry := func() {
				beforeThreadSleepChan <- true
			}
			beforeRetryDelay := func() {
				duration := testClock.Now().Add(testCase.executeDuration[count-1])
				testClock.SetNow(duration)
			}
			timeoutExecutor := NewTimeout(
				logger,
				runtime,
				shortBackOff,
				longBackOff,
				testClock,
				testCase.timeout,
				&beforeRetryDelay,
				&beforeSkipRetry,
			)

			go func() {
				retries, err := timeoutExecutor.WithRetry(ct, execute)
				assert.Equal(t, testCase.expectRetries, retries)
				if testCase.expectErr == nil {
					assert.Nil(t, err)
				} else {
					assert.Equal(t, testCase.expectErr.Code, err.Code)
				}
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
