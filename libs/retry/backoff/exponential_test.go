package backoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/randgen/randgen_test"
)

func TestExponential(t *testing.T) {
	testCases := []struct {
		name             string
		minDelay         time.Duration
		maxDelay         time.Duration
		scalingFactor    int
		randomOffset     int
		randomOffsetUnit time.Duration
		nextDelay        time.Duration
		resetOnSuccess   bool
		randomInts       []int
		expectedDelays   []time.Duration
		executeSucceed   []bool
	}{
		{
			name:             "Without resetOnSuccess",
			minDelay:         2,
			maxDelay:         60,
			scalingFactor:    2,
			randomOffset:     10,
			randomOffsetUnit: time.Nanosecond,
			resetOnSuccess:   false,
			randomInts:       []int{1},
			expectedDelays: []time.Duration{
				5 * time.Nanosecond,
				9 * time.Nanosecond,
				17 * time.Nanosecond,
				33 * time.Nanosecond,
				61 * time.Nanosecond,
				31 * time.Nanosecond,
				16 * time.Nanosecond,
				8 * time.Nanosecond,
				4 * time.Nanosecond,
			},
			executeSucceed: []bool{false, false, false, false, false, true, true, true, true},
		},
		{
			name:             "With resetOnSuccess",
			minDelay:         2,
			maxDelay:         60,
			scalingFactor:    2,
			randomOffset:     10,
			randomOffsetUnit: time.Nanosecond,
			resetOnSuccess:   true,
			randomInts:       []int{1},
			expectedDelays: []time.Duration{
				5 * time.Nanosecond,
				9 * time.Nanosecond,
				17 * time.Nanosecond,
				3 * time.Nanosecond,
				3 * time.Nanosecond,
				3 * time.Nanosecond,
				3 * time.Nanosecond,
			},
			executeSucceed: []bool{false, false, false, true, true, true, true},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			randGen := randgen_test.NewBuiltinRanGen(testCase.randomInts)
			exponential := NewExponentialBuilder().
				MinDelay(testCase.minDelay).
				MaxDelay(testCase.maxDelay).
				RandGenerator(randGen).
				ScalingFactor(testCase.scalingFactor).
				RandomOffset(testCase.randomOffset).
				ResetOnSuccess(testCase.resetOnSuccess).
				RandomOffsetUnit(testCase.randomOffsetUnit).
				Build()

			for index := 0; index < len(testCase.executeSucceed); index++ {
				if testCase.executeSucceed[index] {
					exponential.OnSuccess()
				} else {
					exponential.OnFailure()
				}

				assert.Equal(t, exponential.Delay(), testCase.expectedDelays[index])
			}
		})
	}
}
