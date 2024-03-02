package lang

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanner_Scan(t *testing.T) {
	testCases := []struct {
		name           string
		source         string
		ExpectedTokens []Token
		ExpectedErrs   []Err
	}{
		{
			name:   "empty source",
			source: "",
			ExpectedTokens: []Token{{
				Type:   EOFTokenType,
				Line:   1,
				Column: 1,
			}},
			ExpectedErrs: []Err{},
		},
		{
			name:   "simple String",
			source: `  "Hello World!"   `,
			ExpectedTokens: []Token{
				{
					Type:   StringTokenType,
					Lexeme: "\"Hello World!\"",
					Value:  "Hello World!",
					Line:   1,
					Column: 3,
				},
				{
					Type:   EOFTokenType,
					Line:   1,
					Column: 20,
				},
			},
			ExpectedErrs: []Err{},
		},
		{
			name:   "simple Int",
			source: ` 123 `,
			ExpectedTokens: []Token{
				{
					Type:   IntTokenType,
					Lexeme: "123",
					Value:  int64(123),
					Line:   1,
					Column: 2,
				},
				{
					Type:   EOFTokenType,
					Line:   1,
					Column: 6,
				},
			},
			ExpectedErrs: []Err{},
		},
		{
			name:   "simple Decimal",
			source: ` 123.456 `,
			ExpectedTokens: []Token{
				{
					Type:   DecimalTokenType,
					Lexeme: "123.456",
					Value:  123.456,
					Line:   1,
					Column: 2,
				},
				{
					Type:   EOFTokenType,
					Line:   1,
					Column: 10,
				},
			},
			ExpectedErrs: []Err{},
		},
		{
			name:   "simple Bool",
			source: ` true false `,
			ExpectedTokens: []Token{
				{
					Type:   BoolTokenType,
					Lexeme: "true",
					Value:  true,
					Line:   1,
					Column: 2,
				},
				{
					Type:   BoolTokenType,
					Lexeme: "false",
					Value:  false,
					Line:   1,
					Column: 7,
				},
				{
					Type:   EOFTokenType,
					Line:   1,
					Column: 13,
				},
			},
			ExpectedErrs: []Err{},
		},
		{
			name:   "simple nil",
			source: ` nil `,
			ExpectedTokens: []Token{
				{
					Type:   NilTokenType,
					Lexeme: "nil",
					Line:   1,
					Column: 2,
				},
				{
					Type:   EOFTokenType,
					Line:   1,
					Column: 6,
				},
			},
			ExpectedErrs: []Err{},
		},
		{
			name: "simple code block",
			source: `
				let numOfAdults = 0;
				let age1 = 16;
				
				if (age >= 18) {
					numOfAdults += 1;
				}
				
				let result = (numOfAdults + 10) * 2;
`,
			ExpectedTokens: []Token{
				{
					Type:   LetKeywordTokenType,
					Lexeme: "let",
					Line:   2,
					Column: 5,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "numOfAdults",
					Value:  "numOfAdults",
					Line:   2,
					Column: 9,
				},
				{
					Type:   AssignTokenType,
					Lexeme: "=",
					Line:   2,
					Column: 21,
				},
				{
					Type:   IntTokenType,
					Lexeme: "0",
					Value:  int64(0),
					Line:   2,
					Column: 23,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   2,
					Column: 24,
				},
				{
					Type:   LetKeywordTokenType,
					Lexeme: "let",
					Line:   3,
					Column: 5,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "age1",
					Value:  "age1",
					Line:   3,
					Column: 9,
				},
				{
					Type:   AssignTokenType,
					Lexeme: "=",
					Line:   3,
					Column: 14,
				},
				{
					Type:   IntTokenType,
					Lexeme: "16",
					Value:  int64(16),
					Line:   3,
					Column: 16,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   3,
					Column: 18,
				},
				{
					Type:   IfKeywordTokenType,
					Lexeme: "if",
					Line:   5,
					Column: 5,
				},
				{
					Type:   LeftParenthesisTokenType,
					Lexeme: "(",
					Line:   5,
					Column: 8,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "age",
					Value:  "age",
					Line:   5,
					Column: 9,
				},
				{
					Type:   GreaterThanOrEqualTokenType,
					Lexeme: ">=",
					Line:   5,
					Column: 13,
				},
				{
					Type:   IntTokenType,
					Lexeme: "18",
					Value:  int64(18),
					Line:   5,
					Column: 16,
				},
				{
					Type:   RightParenthesisTokenType,
					Lexeme: ")",
					Line:   5,
					Column: 18,
				},
				{
					Type:   LeftBraceTokenType,
					Lexeme: "{",
					Line:   5,
					Column: 20,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "numOfAdults",
					Value:  "numOfAdults",
					Line:   6,
					Column: 6,
				},
				{
					Type:   AddAssignTokenType,
					Lexeme: "+=",
					Line:   6,
					Column: 18,
				},
				{
					Type:   IntTokenType,
					Lexeme: "1",
					Value:  int64(1),
					Line:   6,
					Column: 21,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   6,
					Column: 22,
				},
				{
					Type:   RightBraceTokenType,
					Lexeme: "}",
					Line:   7,
					Column: 5,
				},
				{
					Type:   LetKeywordTokenType,
					Lexeme: "let",
					Line:   9,
					Column: 5,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "result",
					Value:  "result",
					Line:   9,
					Column: 9,
				},
				{
					Type:   AssignTokenType,
					Lexeme: "=",
					Line:   9,
					Column: 16,
				},
				{
					Type:   LeftParenthesisTokenType,
					Lexeme: "(",
					Line:   9,
					Column: 18,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "numOfAdults",
					Value:  "numOfAdults",
					Line:   9,
					Column: 19,
				},
				{
					Type:   AddTokenType,
					Lexeme: "+",
					Line:   9,
					Column: 31,
				},
				{
					Type:   IntTokenType,
					Lexeme: "10",
					Value:  int64(10),
					Line:   9,
					Column: 33,
				},
				{
					Type:   RightParenthesisTokenType,
					Lexeme: ")",
					Line:   9,
					Column: 35,
				},
				{
					Type:   StarTokenType,
					Lexeme: "*",
					Line:   9,
					Column: 37,
				},
				{
					Type:   IntTokenType,
					Lexeme: "2",
					Value:  int64(2),
					Line:   9,
					Column: 39,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   9,
					Column: 40,
				},
				{
					Type:   EOFTokenType,
					Line:   10,
					Column: 1,
				},
			},
			ExpectedErrs: []Err{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			scanner := NewScanner()
			tokens, errs := scanner.ScanTokens(testCase.source)
			require.Equal(t, testCase.ExpectedTokens, tokens)
			require.Equal(t, testCase.ExpectedErrs, errs)
		})
	}
}
