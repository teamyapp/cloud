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

func TestInfinite(t *testing.T) {
	var transientTimeoutErr errs.ErrorCode = errs.Timeout
	var clientInteractionAlreadyExistsErr errs.ErrorCode = errs.AlreadyExists
	var outageUnimplementedErr errs.ErrorCode = errs.Unimplemented

	testCases := []struct {
		name           string
		errCodes       [](*errs.ErrorCode)
		durations      []time.Duration
		sleepAwakeCount int
		expectRetries     int
		expectErr         *errs.Error
	}{
		{
			name:     "Stop retry with ClientInteraction err category",
			errCodes: []*errs.ErrorCode{
			  &transientTimeoutErr, 
			  &outageUnimplementedErr, 
			  &clientInteractionAlreadyExistsErr,
			},
			durations: []time.Duration{
				401000000, 
				501000000, 
				501000000,
			},
			resRetries:     3,
			resErr:         &errs.Error{Code: errs.InvalidArgument},
			sleepAwakeLoop: 2,
		},
		{
			name:     "Should retry until succeed with Transient and Outage err categories",
			errCodes: []*errs.ErrorCode{&transientTimeoutErr, &outageUnimplementedErr, &transientTimeoutErr, &outageUnimplementedErr, nil},
			durations: []time.Duration{
				401000000, 
				501000000, 
				801000000, 
				1001000000, 
				1001000000,
			},
			resRetries:     5,
			resErr:         nil,
			sleepAwakeLoop: 4,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			randGen := randgen_test.NewBuiltinRanGen([]int{1})
			beforeThreadSleepChan := make(chan bool)
			shortBackOff := backoff.NewExponentialBuilder().RandGenerator(randGen).Build()
			longBackOff := backoff.NewExponentialBuilder().
				RandGenerator(randGen).
				MinDelay(250 * time.Millisecond).
				ResetOnSuccess(true).
				Build()
			var currDuration time.Duration
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
			infiniteExecutor := NewInfinite(telemetry.NewDataCollector(logger), runtime, shortBackOff, longBackOff, &beforeSkipRetry)

			go func() {
				retries, err := infiniteExecutor.WithRetry(ct, execute)

				assert.Equal(t, testCase.resRetries, retries)
				assert.Equal(t, testCase.resErr, err)
			}()

			retry := 1
			for retry <= testCase.sleepAwakeLoop {
				<-beforeThreadSleepChan
				assert.Equal(t, retry, count)
				assert.Equal(t, testCase.durations[retry-1], currDuration)
				runtime.Awake()
				retry++
			}
		})
	}
}
