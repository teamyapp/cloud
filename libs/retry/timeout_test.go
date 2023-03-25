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
	var transientTimeoutErr errs.ErrorCode = errs.Timeout
	var clientInteractionAlreadyExistsErr errs.ErrorCode = errs.AlreadyExists
	var outageUnimplementedErr errs.ErrorCode = errs.Unimplemented
	currentTime := time.Now()
	testClock := runtime_test.NewTestClock(currentTime)
	testCases := []struct {
		name            string
		errCodes        [](*errs.ErrorCode)
		durations       []time.Duration
		timeout         time.Duration
		executeDuration []time.Duration
		resRetries      int
		resErr          *errs.Error
		sleepAwakeLoop  int
	}{
		{
			name:     "Stop retry with ClientInteraction err category",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &clientInteractionAlreadyExistsErr},
			durations: []time.Duration{
				401000000, 501000000, 501000000,
			},
			timeout:         10 * time.Second,
			executeDuration: []time.Duration{2, 3, 4},
			resRetries:      3,
			resErr:          &errs.Error{Code: errs.InvalidArgument},
			sleepAwakeLoop:  2,
		},
		{
			name:     "Succeed before reaching max timeout",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, nil},
			durations: []time.Duration{
				401000000, 501000000, 501000000,
			},
			timeout:         10 * time.Second,
			executeDuration: []time.Duration{2, 3, 4},
			resRetries:      3,
			resErr:          nil,
			sleepAwakeLoop:  2,
		},
		{
			name:     "Not Succeed, max timeout reached",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &transientTimeoutErr},
			durations: []time.Duration{
				401000000, 501000000, 501000000,
			},
			timeout:         10 * time.Second,
			executeDuration: []time.Duration{2 * time.Second, 3 * time.Second, 4 * time.Second},
			resRetries:      3,
			resErr:          &errs.Error{Code: transientTimeoutErr},
			sleepAwakeLoop:  2,
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
				testClock.SetTime(testClock.Now().Add(d))
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
			beforeRetryDelay := func() {
				testClock.SetTime(testClock.Now().Add(testCase.executeDuration[count-1]))
			}
			timeoutExecutor := NewTimeout(
				telemetry.NewDataCollector(logger),
				shortBackOff,
				longBackOff,
				runtime,
				&testClock,
				testCase.timeout,
				&beforeRetryDelay,
				&beforeSkipRetry,
			)

			go func() {
				retries, err := timeoutExecutor.WithRetry(ct, execute)

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
