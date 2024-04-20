package lang

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStaticAnalyzer_Analyze(t *testing.T) {
	testCases := []struct {
		name            string
		source          string
		expectHasErr    bool
		expectErrLine   int
		expectErrColumn int
	}{
		{
			name: "error: duplicate local variable declaration",
			source: `
				func bad() {
				  let name = "first";
				  let name = "second";
				}
			`,
			expectHasErr:    true,
			expectErrLine:   4,
			expectErrColumn: 11,
		},
		{
			name: "error: return at top level",
			source: `
				return 5;
			`,
			expectHasErr:    true,
			expectErrLine:   2,
			expectErrColumn: 5,
		},
		{
			name: "error: local variable is never used",
			source: `
				func bad() {
				  let declared = "first";
				  let defined;
				  let used = 1;
				  used = used + 1;
				}
			`,
			expectHasErr:    true,
			expectErrLine:   3,
			expectErrColumn: 11,
		},
		{
			name: "error: use this outside of class",
			source: `
				func bad() {
				  this;
				}
			`,
			expectHasErr:    true,
			expectErrLine:   3,
			expectErrColumn: 7,
		},
		{
			name: "error: use super outside of class",
			source: `
				func bad() {
				  super.cool();
				}
			`,
			expectHasErr:    true,
			expectErrLine:   3,
			expectErrColumn: 7,
		},
		{
			name: "error: use super outside of sub class",
			source: `
				class Bad {
				  bad() {
					super.cool();
				  }
				}
			`,
			expectHasErr:    true,
			expectErrLine:   4,
			expectErrColumn: 6,
		},
		{
			name: "error: return value inside constructor",
			source: `
				class Bad {
				  constructor() {
					return 5;
				  }
				}
			`,
			expectHasErr:    true,
			expectErrLine:   4,
			expectErrColumn: 6,
		},
		{
			name: "error: class inherit itself",
			source: `
				class Bad : Bad {
				}
			`,
			expectHasErr:    true,
			expectErrLine:   2,
			expectErrColumn: 17,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			scanner := NewScanner()
			tokens, errs := scanner.ScanTokens(tc.source)
			require.Equal(t, 0, len(errs))

			parser := NewParser()
			statements, errs := parser.Parse(tokens)
			require.Equal(t, 0, len(errs))

			s := NewStaticAnalyzer()
			err := s.Analyze(statements)
			if tc.expectHasErr {
				require.NotNil(t, err)
				require.Equal(t, tc.expectErrLine, err.Line)
				require.Equal(t, tc.expectErrColumn, err.Column)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
