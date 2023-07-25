package duration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/libs/errs"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		input            string
		expectedDuration time.Duration
		expectedHasErr   bool
		expectedErrCode  errs.ErrorCode
	}{
		{
			input:           "",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:           "P",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:           "PT",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:           "PT2H1H",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:           "PT2H1S3H",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:           "P2HT1S",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:           "P2D3MT1S",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:           "PT1S2H3M",
			expectedHasErr:  true,
			expectedErrCode: errs.InvalidArgument,
		},
		{
			input:            "P1DT",
			expectedDuration: dayInNanos,
			expectedHasErr:   true,
			expectedErrCode:  errs.InvalidArgument,
		},
		{
			input:            "PT0S",
			expectedDuration: time.Duration(0),
		},
		{
			input:            "PT5S",
			expectedDuration: 5 * time.Second,
		},
		{
			input:            "PT1M",
			expectedDuration: time.Minute,
		},
		{
			input:            "PT5M7S",
			expectedDuration: 5*time.Minute + 7*time.Second,
		},
		{
			input:            "PT1H",
			expectedDuration: time.Hour,
		},
		{
			input:            "PT3H4M",
			expectedDuration: 3*time.Hour + 4*time.Minute,
		},
		{
			input:            "PT3H4M5S",
			expectedDuration: 3*time.Hour + 4*time.Minute + 5*time.Second,
		},
		{
			input:            "P1D",
			expectedDuration: dayInNanos,
		},
		{
			input:            "P1DT3H4M5S",
			expectedDuration: dayInNanos + 3*time.Hour + 4*time.Minute + 5*time.Second,
		},
		{
			input:            "P2Y3M4DT5M",
			expectedDuration: 2*yearInNanos + 3*monthInNanos + 4*dayInNanos + 5*time.Minute,
		},
		{
			input:            "P1Y1M1W1DT1H1M1S",
			expectedDuration: yearInNanos + monthInNanos + weekInNanos + dayInNanos + hourInNanos + minuteInNanos + secondInNanos,
		},
		{
			input:            "-PT0S",
			expectedDuration: time.Duration(0),
		},
		{
			input:            "-PT5S",
			expectedDuration: -5 * time.Second,
		},
		{
			input:            "-PT1M",
			expectedDuration: -time.Minute,
		},
		{
			input:            "-PT5M7S",
			expectedDuration: -5*time.Minute - 7*time.Second,
		},
		{
			input:            "-PT1H",
			expectedDuration: -time.Hour,
		},
		{
			input:            "-PT3H4M",
			expectedDuration: -3*time.Hour - 4*time.Minute,
		},
		{
			input:            "-PT3H4M5S",
			expectedDuration: -3*time.Hour - 4*time.Minute - 5*time.Second,
		},
		{
			input:            "-P1D",
			expectedDuration: -dayInNanos,
		},
		{
			input:            "-P1DT3H4M5S",
			expectedDuration: -dayInNanos - 3*time.Hour - 4*time.Minute - 5*time.Second,
		},
		{
			input:            "-P2Y3M4DT5M",
			expectedDuration: -2*yearInNanos - 3*monthInNanos - 4*dayInNanos - 5*time.Minute,
		},
		{
			input:            "-P1Y1M1W1DT1H1M1S",
			expectedDuration: -yearInNanos - monthInNanos - weekInNanos - dayInNanos - hourInNanos - minuteInNanos - secondInNanos,
		},
	}
	ct := context.Background()
	for _, testCase := range testCases {
		testCase := testCase
		name := testCase.input
		if len(name) == 0 {
			name = "Empty"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			duration, err := Parse(ct, testCase.input)
			if testCase.expectedHasErr {
				require.NotNil(t, err)
				require.Equal(t, testCase.expectedErrCode, err.Code)
				return
			}

			require.Nil(t, err)
			require.Equal(t, testCase.expectedDuration, duration)
		})
	}
}

func TestFormat(t *testing.T) {
	testCases := []struct {
		input                     string
		expectedFormattedDuration string
	}{
		{
			input:                     "PT0S",
			expectedFormattedDuration: "PT0S",
		},
		{
			input:                     "PT5S",
			expectedFormattedDuration: "PT5S",
		},
		{
			input:                     "PT1M",
			expectedFormattedDuration: "PT1M",
		},
		{
			input:                     "PT5M7S",
			expectedFormattedDuration: "PT5M7S",
		},
		{
			input:                     "PT1H",
			expectedFormattedDuration: "PT1H",
		},
		{
			input:                     "PT3H4M",
			expectedFormattedDuration: "PT3H4M",
		},
		{
			input:                     "PT3H4M5S",
			expectedFormattedDuration: "PT3H4M5S",
		},
		{
			input:                     "P1D",
			expectedFormattedDuration: "P1D",
		},
		{
			input:                     "P1DT3H4M5S",
			expectedFormattedDuration: "P1DT3H4M5S",
		},
		{
			input:                     "P1Y1M1W1DT1H1M1S",
			expectedFormattedDuration: "P1Y1M1W1DT1H1M1S",
		},
		{
			input:                     "PT675S",
			expectedFormattedDuration: "PT11M15S",
		},
		{
			input:                     "P29DT70S",
			expectedFormattedDuration: "P4W1DT1M10S",
		},
		{
			input:                     "P31D",
			expectedFormattedDuration: "P1M1D",
		},
		{
			input:                     "P61D",
			expectedFormattedDuration: "P2M1D",
		},
		{
			input:                     "P12M5D",
			expectedFormattedDuration: "P1Y",
		},
		{
			input:                     "P10Y13M36D",
			expectedFormattedDuration: "P11Y2M1D",
		},
		{
			input:                     "P10Y13M36DT0S",
			expectedFormattedDuration: "P11Y2M1D",
		},
		{
			input:                     "-PT0S",
			expectedFormattedDuration: "PT0S",
		},
		{
			input:                     "-PT5S",
			expectedFormattedDuration: "-PT5S",
		},
		{
			input:                     "-PT1M",
			expectedFormattedDuration: "-PT1M",
		},
		{
			input:                     "-PT5M7S",
			expectedFormattedDuration: "-PT5M7S",
		},
		{
			input:                     "-PT1H",
			expectedFormattedDuration: "-PT1H",
		},
		{
			input:                     "-PT3H4M",
			expectedFormattedDuration: "-PT3H4M",
		},
		{
			input:                     "-PT3H4M5S",
			expectedFormattedDuration: "-PT3H4M5S",
		},
		{
			input:                     "-P1D",
			expectedFormattedDuration: "-P1D",
		},
		{
			input:                     "-P1DT3H4M5S",
			expectedFormattedDuration: "-P1DT3H4M5S",
		},
		{
			input:                     "-P1Y1M1W1DT1H1M1S",
			expectedFormattedDuration: "-P1Y1M1W1DT1H1M1S",
		},
		{
			input:                     "-PT675S",
			expectedFormattedDuration: "-PT11M15S",
		},
		{
			input:                     "-P29DT70S",
			expectedFormattedDuration: "-P4W1DT1M10S",
		},
		{
			input:                     "-P31D",
			expectedFormattedDuration: "-P1M1D",
		},
		{
			input:                     "-P61D",
			expectedFormattedDuration: "-P2M1D",
		},
		{
			input:                     "-P12M5D",
			expectedFormattedDuration: "-P1Y",
		},
		{
			input:                     "-P10Y13M36D",
			expectedFormattedDuration: "-P11Y2M1D",
		},
		{
			input:                     "-P10Y13M36DT0S",
			expectedFormattedDuration: "-P11Y2M1D",
		},
	}

	ct := context.Background()
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()
			duration, err := Parse(ct, testCase.input)
			require.Nil(t, err)

			actualFormattedDuration := Format(duration)
			require.Equal(t, testCase.expectedFormattedDuration, actualFormattedDuration)
		})
	}
}
