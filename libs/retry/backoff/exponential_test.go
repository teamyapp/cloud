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
		expectedDelays          []time.Duration
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
			expects:          []time.Duration{5 * time.Nanosecond, 26 * time.Nanosecond, 61 * time.Nanosecond, 8 * time.Nanosecond, 3 * time.Nanosecond},
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
			expects:          []time.Duration{5 * time.Nanosecond, 26 * time.Nanosecond, 61 * time.Nanosecond, 2 * time.Nanosecond, 2 * time.Nanosecond},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {

			exponentialBuilder := NewExponentialBuilder()

			exponentialBuilder = exponentialBuilder.MinDelay(testCase.minDelay).MaxDelay(testCase.maxDelay).RandGenerator(randgen_test.NewBuiltinRanGen(testCase.randomInts)).ScalingFactor(testCase.scalingFactor).RandomOffset(testCase.randomOffset).ResetOnSuccess(testCase.resetOnSuccess).RandomOffsetUnit(testCase.randomOffsetUnit)

			exponential := exponentialBuilder.Build()

			assert.Equal(t, exponential.maxDelay, testCase.maxDelay)
			assert.Equal(t, exponential.minDelay, testCase.minDelay)
			assert.Equal(t, exponential.scalingFactor, testCase.scalingFactor)
			assert.Equal(t, exponential.randomOffset, testCase.randomOffset)
			assert.Equal(t, exponential.resetOnSuccess, testCase.resetOnSuccess)

			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), testCase.expects[0])
			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), testCase.expects[1])
			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), testCase.expects[2])
			exponential.OnSuccess()
			assert.Equal(t, exponential.Delay(), testCase.expects[3])

			exponential.OnSuccess()
			exponential.OnSuccess()
			exponential.OnSuccess()

			assert.Equal(t, exponential.Delay(), testCase.expects[4])
		})
	}
}
