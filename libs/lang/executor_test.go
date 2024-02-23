package lang

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		name           string
		source         string
		expectedValues []any
		expectedOutput string
		expectHasErr   bool
	}{
		{
			name:           "simple nil",
			source:         `nil;`,
			expectedValues: []any{nil},
			expectedOutput: "",
		},
		{
			name:           "simple expression",
			source:         `2 + 5 * 6;`,
			expectedValues: []any{int64(32)},
		},
		{
			name:           "simple expression with grouping",
			source:         `(2 + 5) * 6;`,
			expectedValues: []any{int64(42)},
		},
		{
			name:           "simple expression with unary",
			source:         `!!5;`,
			expectedValues: []any{true},
		},
		{
			name:           "simple equality expression",
			source:         `2 == 5;`,
			expectedValues: []any{false},
		},
		{
			name:           "simple nil equality expression",
			source:         `nil == 2;`,
			expectedValues: []any{false},
		},
		{
			name:           "simple comparison expression",
			source:         `2 > 5;`,
			expectedValues: []any{false},
		},
		{
			name:           "simple equality expression with comparison",
			source:         `2 >= 5 == true;`,
			expectedValues: []any{false},
		},
		{
			name:           "simple comparison expression with term",
			source:         `2 + 5 > 6;`,
			expectedValues: []any{true},
		},
		{
			name:           "simple term expression with factor",
			source:         `2 * 5 + 6;`,
			expectedValues: []any{int64(16)},
		},
		{
			name:           "simple factor expression with unary",
			source:         `2 * -5;`,
			expectedValues: []any{int64(-10)},
		},
		{
			name:           "simple string equality expression equal",
			source:         `"Hello" == "World";`,
			expectedValues: []any{false},
		},
		{
			name:           "simple string equality expression not equal",
			source:         `"Hello" == "Hel" + "lo";`,
			expectedValues: []any{true},
		},
		{
			name:           "simple string concatenation expression",
			source:         `"Hello" + " " + "World";`,
			expectedValues: []any{"Hello World"},
		},
		{
			name:           "complex expression",
			source:         `!(2 * (5 + 6) <= 4) == true;`,
			expectedValues: []any{true},
		},
		{
			name:           "simple bitwise left shift expression",
			source:         `2 << 2;`,
			expectedValues: []any{int64(8)},
		},
		{
			name:           "simple bitwise right shift expression",
			source:         `8 >> 2;`,
			expectedValues: []any{int64(2)},
		},
		{
			name:           "simple bitwise or expression",
			source:         `2 | 4;`,
			expectedValues: []any{int64(6)},
		},
		{
			name:           "simple bitwise xor expression",
			source:         `2 ^ 4;`,
			expectedValues: []any{int64(6)},
		},
		{
			name:           "simple bitwise and expression",
			source:         `2 & 4;`,
			expectedValues: []any{int64(0)},
		},
		{
			name:           "simple division expression",
			source:         `10 / 2;`,
			expectedValues: []any{int64(5)},
		},
		{
			name:           "simple division expression with float",
			source:         `10.0 / 3.0;`,
			expectedValues: []any{float64(10) / float64(3)},
		},
		{
			name:           "simple modulo expression",
			source:         `10 % 3;`,
			expectedValues: []any{int64(1)},
		},
		{
			name:           "simple ternary expression",
			source:         `true ? 2 : 3;`,
			expectedValues: []any{int64(2)},
		},
		{
			name:           "complex ternary expression",
			source:         `1 == 2 ? 2 + 5 * 6 : 3 + 4 / 2;`,
			expectedValues: []any{int64(5)},
		},
		{
			name: "multiple statements",
			source: `
				2 + 5 * 6; 
				2 == 5;`,
			expectedValues: []any{int64(32), false},
		},
		{
			name: "define and read variables",
			source: `
				let a = 5;
				let b = 6;
				let c = a + b;
				c * 2;`,
			expectedValues: []any{int64(22)},
		},
		{
			name: "define, assign value to and read from variables",
			source: `
				let a = 5;
				let b = 6;
				a = 1 + 2 * b;
				b = a - 4;
`,
			expectedValues: []any{int64(13), int64(9)},
		},
		{
			name: "variable scope with block statement",
			source: `
				let a = 5;
				let b = 6;
				{
					let a = 8;
					b = b + 1;
					print(a + " " + b);
				}
				print(" | ");
				print(a + " " + b);
			`,
			expectedOutput: "8 7 | 5 7",
		},
		{
			name: "access uninitialized variables",
			source: `
				let a;
				let b;
				let c = 1;
				b = 10;
				c;
				a;
			`,
			expectHasErr:   true,
			expectedValues: []any{int64(10), int64(1)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := NewScanner()
			tokens, errs := scanner.ScanTokens(tc.source)
			require.Equal(t, 0, len(errs))

			parser := NewParser()
			statements, errs := parser.Parse(tokens)
			require.Equal(t, 0, len(errs))

			environment := NewEnvironment()
			var outputBuf bytes.Buffer
			executor := NewExecutor(environment, &outputBuf)
			values, err := executor.Execute(statements)
			if tc.expectHasErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, tc.expectedValues, values)
			require.Equal(t, tc.expectedOutput, outputBuf.String())
		})
	}
}
