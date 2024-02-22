package lang

import (
	"github.com/stretchr/testify/require"
	"testing"
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

let result = (numOfAdults + 10) * 2;`,
			ExpectedTokens: []Token{
				{
					Type:   LetKeywordTokenType,
					Lexeme: "let",
					Line:   2,
					Column: 1,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "numOfAdults",
					Value:  "numOfAdults",
					Line:   2,
					Column: 5,
				},
				{
					Type:   AssignTokenType,
					Lexeme: "=",
					Line:   2,
					Column: 17,
				},
				{
					Type:   IntTokenType,
					Lexeme: "0",
					Value:  int64(0),
					Line:   2,
					Column: 19,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   2,
					Column: 20,
				},
				{
					Type:   LetKeywordTokenType,
					Lexeme: "let",
					Line:   3,
					Column: 1,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "age1",
					Value:  "age1",
					Line:   3,
					Column: 5,
				},
				{
					Type:   AssignTokenType,
					Lexeme: "=",
					Line:   3,
					Column: 10,
				},
				{
					Type:   IntTokenType,
					Lexeme: "16",
					Value:  int64(16),
					Line:   3,
					Column: 12,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   3,
					Column: 14,
				},
				{
					Type:   IfKeywordTokenType,
					Lexeme: "if",
					Line:   5,
					Column: 1,
				},
				{
					Type:   LeftParenthesisTokenType,
					Lexeme: "(",
					Line:   5,
					Column: 4,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "age",
					Value:  "age",
					Line:   5,
					Column: 5,
				},
				{
					Type:   GreaterThanOrEqualTokenType,
					Lexeme: ">=",
					Line:   5,
					Column: 9,
				},
				{
					Type:   IntTokenType,
					Lexeme: "18",
					Value:  int64(18),
					Line:   5,
					Column: 12,
				},
				{
					Type:   RightParenthesisTokenType,
					Lexeme: ")",
					Line:   5,
					Column: 14,
				},
				{
					Type:   LeftBraceTokenType,
					Lexeme: "{",
					Line:   5,
					Column: 16,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "numOfAdults",
					Value:  "numOfAdults",
					Line:   6,
					Column: 5,
				},
				{
					Type:   AddAssignTokenType,
					Lexeme: "+=",
					Line:   6,
					Column: 17,
				},
				{
					Type:   IntTokenType,
					Lexeme: "1",
					Value:  int64(1),
					Line:   6,
					Column: 20,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   6,
					Column: 21,
				},
				{
					Type:   RightBraceTokenType,
					Lexeme: "}",
					Line:   7,
					Column: 1,
				},
				{
					Type:   LetKeywordTokenType,
					Lexeme: "let",
					Line:   9,
					Column: 1,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "result",
					Value:  "result",
					Line:   9,
					Column: 5,
				},
				{
					Type:   AssignTokenType,
					Lexeme: "=",
					Line:   9,
					Column: 12,
				},
				{
					Type:   LeftParenthesisTokenType,
					Lexeme: "(",
					Line:   9,
					Column: 14,
				},
				{
					Type:   IdentifierTokenType,
					Lexeme: "numOfAdults",
					Value:  "numOfAdults",
					Line:   9,
					Column: 15,
				},
				{
					Type:   AddTokenType,
					Lexeme: "+",
					Line:   9,
					Column: 27,
				},
				{
					Type:   IntTokenType,
					Lexeme: "10",
					Value:  int64(10),
					Line:   9,
					Column: 29,
				},
				{
					Type:   RightParenthesisTokenType,
					Lexeme: ")",
					Line:   9,
					Column: 31,
				},
				{
					Type:   StarTokenType,
					Lexeme: "*",
					Line:   9,
					Column: 33,
				},
				{
					Type:   IntTokenType,
					Lexeme: "2",
					Value:  int64(2),
					Line:   9,
					Column: 35,
				},
				{
					Type:   SemicolonTokenType,
					Lexeme: ";",
					Line:   9,
					Column: 36,
				},
				{
					Type:   EOFTokenType,
					Line:   9,
					Column: 37,
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
