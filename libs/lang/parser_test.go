package lang

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	testCases := []struct {
		name               string
		source             string
		expectedExpression string
		hasError           bool
	}{
		{
			name:               "simple expression",
			source:             `2 + 5 * 6`,
			expectedExpression: "(+ 2 (* 5 6))",
		},
		{
			name:               "simple expression with grouping",
			source:             `(2 + 5) * 6`,
			expectedExpression: "(* (group (+ 2 5)) 6)",
		},
		{
			name:               "simple expression with unary",
			source:             `!!true`,
			expectedExpression: "(! (! true))",
		},
		{
			name:               "simple equality expression",
			source:             `2 == 5`,
			expectedExpression: "(== 2 5)",
		},
		{
			name:               "simple comparison expression",
			source:             `2 > 5`,
			expectedExpression: "(> 2 5)",
		},
		{
			name:               "simple equality expression with comparison",
			source:             `2 >= 5 == true`,
			expectedExpression: "(== (>= 2 5) true)",
		},
		{
			name:               "simple comparison expression with term",
			source:             `2 + 5 > 6`,
			expectedExpression: "(> (+ 2 5) 6)",
		},
		{
			name:               "simple term expression with factor",
			source:             `2 * 5 + 6`,
			expectedExpression: "(+ (* 2 5) 6)",
		},
		{
			name:               "simple factor expression with unary",
			source:             `2 * -5`,
			expectedExpression: "(* 2 (- 5))",
		},
		{
			name:               "simple string equality expression",
			source:             `"Hello World!" == "Hello" + " " + "World!"`,
			expectedExpression: `(== "Hello World!" (+ (+ "Hello" " ") "World!"))`,
		},
		{
			name:               "complex expression",
			source:             `!(2 * (5 + 6) <= 4) == true`,
			expectedExpression: "(== (! (group (<= (* 2 (group (+ 5 6))) 4))) true)",
		},
		{
			name:     "expression with unmatched (",
			source:   `!(2 * (5 + 6 <= 4) == true`,
			hasError: true,
		},
		{
			name:               "expressions separated by comma",
			source:             `2 + 5, 6 * 7, 8 == 9, !true, "Hello"`,
			expectedExpression: `(expressionList (+ 2 5) (* 6 7) (== 8 9) (! true) "Hello")`,
		},
		{
			name:               "simple ternary expression",
			source:             `true ? 2 : 3`,
			expectedExpression: "(?: true 2 3)",
		},
		{
			name:               "complex ternary expression",
			source:             `1 == 2 ? 2 + 5 * 6 : 3 + 4 / 2`,
			expectedExpression: "(?: (== 1 2) (+ 2 (* 5 6)) (+ 3 (/ 4 2)))",
		},
		{
			name:               "chained ternary expression",
			source:             `1 == 2 ? 2 + 5 * 6 : 3 + 4 / 2 ? 1 : 2`,
			expectedExpression: "(?: (== 1 2) (+ 2 (* 5 6)) (?: (+ 3 (/ 4 2)) 1 2))",
		},
		{
			name:     "ternary expression with unmatched :",
			source:   "1 == 2 ? 2 + 5 * 6",
			hasError: true,
		},
		{
			name:     "add without left operand",
			source:   `+ 5`,
			hasError: true,
		},
		{
			name:     "add without right operand",
			source:   `5 +`,
			hasError: true,
		},
		{
			name:     "unary without operand",
			source:   `!`,
			hasError: true,
		},
		{
			name:               "simple && and ||",
			source:             `true || false && true`,
			expectedExpression: "(|| true (&& false true))",
		},
		{
			name:               "simple &, | and ^",
			source:             `1 | 2 ^ 3 & 4`,
			expectedExpression: "(| 1 (^ 2 (& 3 4)))",
		},
		{
			name:               "simple << and >>",
			source:             `1 << 2 >> 3`,
			expectedExpression: "(>> (<< 1 2) 3)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens, errs := NewScanner().ScanTokens(tc.source)
			require.Equal(t, 0, len(errs))

			parser := NewParser()
			expr, _ := parser.parse(tokens)
			if tc.hasError {
				require.True(t, len(parser.GetErrors()) > 0)
				return
			}

			require.Equal(t, tc.expectedExpression, expr.String())
		})
	}
}
