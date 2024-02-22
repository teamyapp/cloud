package lang

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		name          string
		source        string
		expectedValue any
		expectHasErr  bool
	}{
		{
			name:          "simple nil",
			source:        `nil`,
			expectedValue: nil,
		},
		{
			name:          "simple expression",
			source:        `2 + 5 * 6`,
			expectedValue: int64(32),
		},
		{
			name:          "simple expression with grouping",
			source:        `(2 + 5) * 6`,
			expectedValue: int64(42),
		},
		{
			name:          "simple expression with unary",
			source:        `!!true`,
			expectedValue: true,
		},
		{
			name:          "simple equality expression",
			source:        `2 == 5`,
			expectedValue: false,
		},
		{
			name:          "simple nil equality expression",
			source:        `nil == 2`,
			expectedValue: false,
		},
		{
			name:          "simple comparison expression",
			source:        `2 > 5`,
			expectedValue: false,
		},
		{
			name:          "simple equality expression with comparison",
			source:        `2 >= 5 == true`,
			expectedValue: false,
		},
		{
			name:          "simple comparison expression with term",
			source:        `2 + 5 > 6`,
			expectedValue: true,
		},
		{
			name:          "simple term expression with factor",
			source:        `2 * 5 + 6`,
			expectedValue: int64(16),
		},
		{
			name:          "simple factor expression with unary",
			source:        `2 * -5`,
			expectedValue: int64(-10),
		},
		{
			name:          "simple string equality expression equal",
			source:        `"Hello" == "World"`,
			expectedValue: false,
		},
		{
			name:          "simple string equality expression not equal",
			source:        `"Hello" == "Hel" + "lo"`,
			expectedValue: true,
		},
		{
			name:          "simple string concatenation expression",
			source:        `"Hello" + " " + "World"`,
			expectedValue: "Hello World",
		},
		{
			name:          "complex expression",
			source:        `!(2 * (5 + 6) <= 4) == true`,
			expectedValue: true,
		},
		{
			name:          "simple bitwise left shift expression",
			source:        `2 << 2`,
			expectedValue: int64(8),
		},
		{
			name:          "simple bitwise right shift expression",
			source:        `8 >> 2`,
			expectedValue: int64(2),
		},
		{
			name:          "simple bitwise or expression",
			source:        `2 | 4`,
			expectedValue: int64(6),
		},
		{
			name:          "simple bitwise xor expression",
			source:        `2 ^ 4`,
			expectedValue: int64(6),
		},
		{
			name:          "simple bitwise and expression",
			source:        `2 & 4`,
			expectedValue: int64(0),
		},
		{
			name:          "simple division expression",
			source:        `10 / 2`,
			expectedValue: int64(5),
		},
		{
			name:          "simple division expression with float",
			source:        `10.0 / 3.0`,
			expectedValue: float64(10) / float64(3),
		},
		{
			name:          "simple modulo expression",
			source:        `10 % 3`,
			expectedValue: int64(1),
		},
		{
			name:          "simple ternary expression",
			source:        `true ? 2 : 3`,
			expectedValue: int64(2),
		},
		{
			name:          "complex ternary expression",
			source:        `1 == 2 ? 2 + 5 * 6 : 3 + 4 / 2`,
			expectedValue: int64(5),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := NewScanner()
			tokens, errs := scanner.ScanTokens(tc.source)
			require.Equal(t, 0, len(errs))

			parser := NewParser()
			expression, err := parser.parse(tokens)
			require.Nil(t, err)

			value, err := Evaluate(expression)
			if tc.expectHasErr {
				require.NotNil(t, err)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tc.expectedValue, value)
		})
	}
}
