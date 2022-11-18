package backoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	testCases := []testCase{
		{
			name:             "initialize all options",
			minDelay:         100 * time.Millisecond,
			maxDelay:         60000 * time.Millisecond,
			scalingFactor:    4,
			randomOffset:     100,
			randomOffsetUnit: time.Millisecond,
			resetOnSuccess:   false,
			randomInts:       []int{},
		},
	}
	
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			exponential := getExponential(testCase)
			
			assert.Equal(t, exponential.maxDelay, testCase.maxDelay)
			assert.Equal(t, exponential.minDelay, testCase.minDelay)
			assert.Equal(t, exponential.scalingFactor, testCase.scalingFactor)
			assert.Equal(t, exponential.randomOffset, testCase.randomOffset)
			assert.Equal(t, exponential.resetOnSuccess, testCase.resetOnSuccess)
		})
	}
}

func TestOnRandom(t *testing.T) {
	testCases := []testCase{
		{
			name:             "test with randomGen",
			minDelay:         2 * time.Millisecond,
			maxDelay:         60000 * time.Millisecond,
			scalingFactor:    2,
			randomOffset:     100,
			randomOffsetUnit: time.Millisecond,
			resetOnSuccess:   false,
			randomInts:       []int{1},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			exponential := getExponential(testCase)

			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), 5*time.Millisecond)
			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), 26*time.Millisecond)
			exponential.OnSuccess()
			assert.Equal(t, exponential.Delay(), 6099019*time.Nanosecond)
		})
	}

}

func TestOnFailure(t *testing.T) {
	testCases := []testCase{
		{
			name:             "onFailure test",
			minDelay:         2 * time.Millisecond,
			maxDelay:         1000 * time.Millisecond,
			scalingFactor:    2,
			randomOffset:     100,
			randomOffsetUnit: time.Millisecond,
			resetOnSuccess:   false,
			randomInts:       []int{},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			exponential := getExponential(testCase)

			assert.Equal(t, exponential.Delay(), testCase.minDelay)
			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), 4*time.Millisecond)
			exponential.OnFailure()
			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), 256*time.Millisecond)

			exponential.OnFailure()
			assert.Equal(t, exponential.Delay(), testCase.maxDelay)
		})
	}
}

func TestOnSuccess(t *testing.T) {
	testCases := []testCase{
		{
			name:             "onSuccess test without resetOnSuccess",
			minDelay:         5 * time.Millisecond,
			maxDelay:         60000 * time.Millisecond,
			scalingFactor:    2,
			randomOffset:     100,
			randomOffsetUnit: time.Millisecond,
			resetOnSuccess:   false,
			randomInts:       []int{},
		},
		{
			name:             "onSuccess test with resetOnSuccess",
			minDelay:         5 * time.Millisecond,
			maxDelay:         60000 * time.Millisecond,
			scalingFactor:    2,
			randomOffset:     200,
			randomOffsetUnit: time.Millisecond,
			resetOnSuccess:   true,
			randomInts:       []int{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			exponential := getExponential(testCase)

			assert.Equal(t, exponential.Delay(), testCase.minDelay)
			exponential.OnFailure()
			exponential.OnFailure()
			exponential.OnSuccess()

			if testCase.resetOnSuccess {
				assert.Equal(t, exponential.Delay(), testCase.minDelay)
			} else {
				assert.Equal(t, exponential.Delay(), 25*time.Millisecond)
				exponential.OnSuccess()
				exponential.OnSuccess()
				exponential.OnSuccess()
				assert.Equal(t, exponential.Delay(), testCase.minDelay)
			}
		})
	}
}

type testCase struct {
	name             string
	minDelay         time.Duration
	maxDelay         time.Duration
	scalingFactor    int
	randomOffset     int
	randomOffsetUnit time.Duration
	nextDelay        time.Duration
	resetOnSuccess   bool
	randomInts       []int
}

type testRandGenerator struct {
	randomInts         []int
	nextRandomIntIndex int
}

func (b testRandGenerator) Intn(i int) int {
	if b.nextRandomIntIndex == len(b.randomInts) {
		b.nextRandomIntIndex = 0
		return 0
	}

	value := b.randomInts[b.nextRandomIntIndex]
	b.nextRandomIntIndex++
	return value
}

func getExponential(testCase testCase) Exponential {
	exponentialBuilder := NewExponentialBuilder()

	exponentialBuilder.minDelay = testCase.minDelay
	exponentialBuilder.maxDelay = testCase.maxDelay
	exponentialBuilder.randGenerator = testRandGenerator{
		randomInts: testCase.randomInts,
	}
	exponentialBuilder.scalingFactor = testCase.scalingFactor
	exponentialBuilder.randomOffset = testCase.randomOffset
	exponentialBuilder.resetOnSuccess = testCase.resetOnSuccess

	exponential := exponentialBuilder.Build()
	return exponential
}
