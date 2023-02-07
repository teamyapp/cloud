package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen/randgen_test"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime/runtime_test"
)

func TestTimeout(t *testing.T) {

	testCases := []struct {
		name             string
		timeout          time.Duration
		err              *errs.Error
		awaits           int
		randomInts       []int
		minDelay         time.Duration
		maxDelay         time.Duration
		scalingFactor    int
		randomOffset     int
		randomOffsetUnit time.Duration
	}{

		{
			name:             "Should succeed before reaching timeout",
			timeout:          19,
			err:              nil,
			awaits:           2,
			randomInts:       []int{1},
			minDelay:         2,
			maxDelay:         60,
			scalingFactor:    2,
			randomOffset:     10,
			randomOffsetUnit: time.Nanosecond,
		},
		{
			name:             "Should get error when reaching timeout",
			timeout:          18,
			err:              &errs.Error{Code: errs.Serialization},
			awaits:           1,
			randomInts:       []int{1},
			minDelay:         2,
			maxDelay:         60,
			scalingFactor:    2,
			randomOffset:     10,
			randomOffsetUnit: time.Nanosecond,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			currentTime := time.Now()
			testClock := runtime_test.NewTestClock(currentTime)
			beforeThreadSleepChan := make(chan bool)
			randGen := randgen_test.NewBuiltinRanGen(testCase.randomInts)
			backoff := backoff.NewExponentialBuilder().
				RandGenerator(randGen).
				MinDelay(testCase.minDelay).
				MaxDelay(testCase.maxDelay).
				ScalingFactor(testCase.scalingFactor).
				RandomOffset(testCase.randomOffset).
				RandomOffsetUnit(testCase.randomOffsetUnit).
				Build()
			builtinRuntime := runtime_test.NewTestRuntime(func(duration time.Duration) {
				testClock.SetTime(testClock.Now().Add(duration))
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
			timeoutExecutor := NewTimeout(builtinRuntime, backoff, &testClock, testCase.timeout, func() {
				testClock.SetTime(testClock.Now().Add(2))
			})

			go func() {
				retries, err := timeoutExecutor.WithRetry(execute)

				assert.Equal(t, 2, retries)
				assert.Equal(t, testCase.err, err)
			}()

			for await := 0; await < testCase.awaits; await++ {
				<-beforeThreadSleepChan
				assert.Equal(t, await+1, count)
				builtinRuntime.Awake()
			}
		})
	}
}
