package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		name   string
		input  Expression
		output Expression
	}{
		{
			name: "simple string",
			input: Expression{
				Type:      ValueExpressionType,
				ValueType: StringDataType,
				Value:     "John",
			},
			output: Expression{
				Type:      ValueExpressionType,
				ValueType: StringDataType,
				Value:     "John",
			},
		},
		{
			name: "function invocation GreaterThan false",
			input: Expression{
				Type: InvocationExpressionType,
				FuncExpression: &Expression{
					Type:       IdentifierExpressionType,
					Identifier: "GreaterThan",
				},
				FuncInputValues: []Expression{
					{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     1,
					},
					{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     2,
					},
				},
			},
			output: Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     false,
			},
		},
		{
			name: "function invocation GreaterThan true",
			input: Expression{
				Type: InvocationExpressionType,
				FuncExpression: &Expression{
					Type:       IdentifierExpressionType,
					Identifier: "GreaterThan",
				},
				FuncInputValues: []Expression{
					{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     2,
					},
					{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     1,
					},
				},
			},
			output: Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     true,
			},
		},
		{
			name: "numbers Equal",
			input: Expression{
				Type: InvocationExpressionType,
				FuncExpression: &Expression{
					Type:       IdentifierExpressionType,
					Identifier: "Equal",
				},
				FuncInputValues: []Expression{
					{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     2,
					},
					{
						Type:      ValueExpressionType,
						ValueType: DecimalDataType,
						Value:     2.0,
					},
				},
			},
			output: Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     true,
			},
		},
		{
			name: "string Contains",
			input: Expression{
				Type: InvocationExpressionType,
				FuncExpression: &Expression{
					Type:       IdentifierExpressionType,
					Identifier: "Contains",
				},
				FuncInputValues: []Expression{
					{
						Type:      ValueExpressionType,
						ValueType: StringDataType,
						Value:     "Hello World!",
					},
					{
						Type:      ValueExpressionType,
						ValueType: StringDataType,
						Value:     "World",
					},
				},
			},
			output: Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     true,
			},
		},
		{
			name: "string Not Contains",
			input: Expression{
				Type: InvocationExpressionType,
				FuncExpression: &Expression{
					Type:       IdentifierExpressionType,
					Identifier: "Not",
				},
				FuncInputValues: []Expression{
					{
						Type: InvocationExpressionType,
						FuncExpression: &Expression{
							Type:       IdentifierExpressionType,
							Identifier: "Contains",
						},
						FuncInputValues: []Expression{
							{
								Type:      ValueExpressionType,
								ValueType: StringDataType,
								Value:     "Hello World!",
							},
							{
								Type:      ValueExpressionType,
								ValueType: StringDataType,
								Value:     "Cool",
							},
						},
					},
				},
			},
			output: Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     true,
			},
		},
		{
			name: "complex expression",
			input: Expression{
				Type: InvocationExpressionType,
				FuncExpression: &Expression{
					Type:       IdentifierExpressionType,
					Identifier: "And",
				},
				FuncInputValues: []Expression{
					{
						Type: InvocationExpressionType,
						FuncExpression: &Expression{
							Type:       IdentifierExpressionType,
							Identifier: "GreaterThan",
						},
						FuncInputValues: []Expression{
							{
								Type:       IdentifierExpressionType,
								Identifier: "person1.Age",
							},
							{
								Type:       IdentifierExpressionType,
								Identifier: "person2.Age",
							},
						},
					},
					{
						Type:      ValueExpressionType,
						ValueType: BoolDataType,
						Value:     true,
					},
				},
			},
			output: Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			env := MapEnvironment(map[string]Expression{})
			UseStd(env)
			env.Set("age", Expression{
				Type:      ValueExpressionType,
				ValueType: IntDataType,
				Value:     10,
			})
			env.Set("person1", Expression{
				Type:      ValueExpressionType,
				ValueType: StructDataType,
				Value: struct {
					FirstName string
					LastName  string
					Age       int
				}{
					FirstName: "Charlotte",
					LastName:  "Doe",
					Age:       10,
				},
			})
			env.Set("person2", Expression{
				Type:      ValueExpressionType,
				ValueType: StructDataType,
				Value: struct {
					FirstName string
					LastName  string
					Age       int
				}{
					FirstName: "John",
					LastName:  "Doe",
					Age:       5,
				},
			})
			output, err := Evaluate(env, testCase.input)
			assert.Nil(t, err)
			assert.Equal(t, testCase.output, output)
		})
	}
}
