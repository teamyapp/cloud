package query

import (
	"fmt"
	"strings"
)

type Ordered interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func UseStd(environment Environment) {
	UseGreaterThan(environment)
	UseLessThan(environment)
	UseAdd(environment)
	UseSubtract(environment)
	UseMultiply(environment)
	UseDivide(environment)
	UseAnd(environment)
	UseOr(environment)
	UseNot(environment)
	UseEqual(environment)
	UseContains(environment)
	UseContainsIgnoreCase(environment)
}

func UseGreaterThan(environment Environment) {
	environment.Set("GreaterThan", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			return withTwoNums(
				inputs,
				func(num1 int64, num2 int64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: BoolDataType,
						Value:     num1 > num2,
					}, nil
				},
				func(num1 float64, num2 float64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: BoolDataType,
						Value:     num1 > num2,
					}, nil
				})
		}),
	})
}

func UseLessThan(environment Environment) {
	environment.Set("GreaterThan", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			return withTwoNums(
				inputs,
				func(num1 int64, num2 int64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: BoolDataType,
						Value:     num1 > num2,
					}, nil
				},
				func(num1 float64, num2 float64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: BoolDataType,
						Value:     num1 > num2,
					}, nil
				})
		}),
	})
}

func UseAdd(environment Environment) {
	environment.Set("Add", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			return withTwoNums(
				inputs,
				func(num1 int64, num2 int64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     num1 + num2,
					}, nil
				},
				func(num1 float64, num2 float64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: DecimalDataType,
						Value:     num1 + num2,
					}, nil
				})
		}),
	})
}

func UseSubtract(environment Environment) {
	environment.Set("Subtract", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			return withTwoNums(
				inputs,
				func(num1 int64, num2 int64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     num1 - num2,
					}, nil
				},
				func(num1 float64, num2 float64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: DecimalDataType,
						Value:     num1 - num2,
					}, nil
				})
		}),
	})
}

func UseMultiply(environment Environment) {
	environment.Set("Multiply", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			return withTwoNums(
				inputs,
				func(num1 int64, num2 int64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     num1 * num2,
					}, nil
				},
				func(num1 float64, num2 float64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: DecimalDataType,
						Value:     num1 * num2,
					}, nil
				})
		}),
	})
}

func UseDivide(environment Environment) {
	environment.Set("Divide", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			return withTwoNums(
				inputs,
				func(num1 int64, num2 int64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: IntDataType,
						Value:     num1 / num2,
					}, nil
				},
				func(num1 float64, num2 float64) (Expression, error) {
					return Expression{
						Type:      ValueExpressionType,
						ValueType: DecimalDataType,
						Value:     num1 / num2,
					}, nil
				})
		}),
	})
}

func UseAnd(environment Environment) {
	environment.Set("And", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			result := true
			for _, input := range inputs {
				if input.Type != ValueExpressionType {
					return Expression{}, fmt.Errorf("input must be a value")
				}

				if input.ValueType != BoolDataType {
					return Expression{}, fmt.Errorf("input must be a bool")
				}

				result = result && input.Value.(bool)
			}

			return Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     result,
			}, nil
		}),
	})
}

func UseOr(environment Environment) {
	environment.Set("Or", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			result := false
			for _, input := range inputs {
				if input.Type != ValueExpressionType {
					return Expression{}, fmt.Errorf("input must be a value")
				}

				if input.ValueType != BoolDataType {
					return Expression{}, fmt.Errorf("input must be a bool")
				}

				result = result || input.Value.(bool)
			}

			return Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     result,
			}, nil
		}),
	})
}

func UseNot(environment Environment) {
	environment.Set("Not", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			if len(inputs) != 1 {
				return Expression{}, fmt.Errorf("Not function requires 1 input")
			}

			if inputs[0].Type != ValueExpressionType {
				return Expression{}, fmt.Errorf("input must be a value")
			}

			if inputs[0].ValueType != BoolDataType {
				return Expression{}, fmt.Errorf("input must be a bool")
			}

			return Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     !inputs[0].Value.(bool),
			}, nil
		}),
	})
}

func UseEqual(environment Environment) {
	environment.Set("Equal", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			if len(inputs) != 2 {
				return Expression{}, fmt.Errorf("EqualTo function requires 2 inputs")
			}

			if inputs[0].Type != ValueExpressionType {
				return Expression{}, fmt.Errorf("input[1] must be a value")
			}

			if inputs[1].Type != ValueExpressionType {
				return Expression{}, fmt.Errorf("input[2] must be a value")
			}

			if inputs[0].ValueType == inputs[1].ValueType {
				return Expression{
					Type:      ValueExpressionType,
					ValueType: BoolDataType,
					Value:     inputs[0].Value == inputs[1].Value,
				}, nil
			}

			return withTwoNums(inputs, func(num1 int64, num2 int64) (Expression, error) {
				return Expression{
					Type:      ValueExpressionType,
					ValueType: BoolDataType,
					Value:     num1 == num2,
				}, nil
			}, func(num1 float64, num2 float64) (Expression, error) {
				return Expression{
					Type:      ValueExpressionType,
					ValueType: BoolDataType,
					Value:     num1 == num2,
				}, nil
			})
		}),
	})
}

func UseContains(environment Environment) {
	environment.Set("Contains", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			if len(inputs) != 2 {
				return Expression{}, fmt.Errorf("Contains function requires 2 inputs")
			}

			if inputs[0].Type != ValueExpressionType {
				return Expression{}, fmt.Errorf("input[1] must be a value")
			}

			if inputs[0].ValueType != StringDataType {
				return Expression{}, fmt.Errorf("input[1] must be a string")
			}

			if inputs[1].Type != ValueExpressionType {
				return Expression{}, fmt.Errorf("input[2] must be a value")
			}

			if inputs[1].ValueType != StringDataType {
				return Expression{}, fmt.Errorf("input[2] must be a string")
			}

			return Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     strings.Contains(inputs[0].Value.(string), inputs[1].Value.(string)),
			}, nil
		}),
	})
}

func UseContainsIgnoreCase(environment Environment) {
	environment.Set("ContainsIgnoreCase", Expression{
		Type:      ValueExpressionType,
		ValueType: FunctionDataType,
		Value: Function(func(inputs ...Expression) (Expression, error) {
			if len(inputs) != 2 {
				return Expression{}, fmt.Errorf("Contains function requires 2 inputs")
			}

			if inputs[0].Type != ValueExpressionType {
				return Expression{}, fmt.Errorf("input[1] must be a value")
			}

			if inputs[0].ValueType != StringDataType {
				return Expression{}, fmt.Errorf("input[1] must be a string")
			}

			if inputs[1].Type != ValueExpressionType {
				return Expression{}, fmt.Errorf("input[2] must be a value")
			}

			if inputs[1].ValueType != StringDataType {
				return Expression{}, fmt.Errorf("input[2] must be a string")
			}

			text := strings.ToLower(inputs[0].Value.(string))
			pattern := strings.ToLower(inputs[0].Value.(string))
			return Expression{
				Type:      ValueExpressionType,
				ValueType: BoolDataType,
				Value:     strings.Contains(text, pattern),
			}, nil
		}),
	})
}

func withTwoNums(
	inputs []Expression,
	ints func(num1 int64, num2 int64) (Expression, error),
	decimals func(num1 float64, num2 float64) (Expression, error),
) (Expression, error) {
	if len(inputs) != 2 {
		return Expression{}, fmt.Errorf("GreaterThan function requires 2 inputs")
	}

	if inputs[0].Type != ValueExpressionType {
		return Expression{}, fmt.Errorf("input[1] must be a value")
	}

	if inputs[0].ValueType != IntDataType && inputs[0].ValueType != DecimalDataType {
		return Expression{}, fmt.Errorf("input[1] must be int and decimal")
	}

	if inputs[1].Type != ValueExpressionType {
		return Expression{}, fmt.Errorf("input[2] must be a value")
	}

	if inputs[1].ValueType != IntDataType && inputs[1].ValueType != DecimalDataType {
		return Expression{}, fmt.Errorf("input[2] must be int and decimal")
	}

	if inputs[0].ValueType == IntDataType && inputs[1].ValueType == IntDataType {
		num1, err := toInt64(inputs[0].Value)
		if err != nil {
			return Expression{}, err
		}

		num2, err := toInt64(inputs[1].Value)
		if err != nil {
			return Expression{}, err
		}

		return ints(num1, num2)
	}

	if inputs[0].ValueType == DecimalDataType && inputs[1].ValueType == DecimalDataType {
		num1, err := toFloat64(inputs[0].Value)
		if err != nil {
			return Expression{}, err
		}

		num2, err := toFloat64(inputs[1].Value)
		if err != nil {
			return Expression{}, err
		}

		return decimals(num1, num2)
	}

	if inputs[0].ValueType == IntDataType {
		num1, err := toInt64(inputs[0].Value)
		if err != nil {
			return Expression{}, err
		}

		num2, err := toFloat64(inputs[1].Value)
		if err != nil {
			return Expression{}, err
		}

		return decimals(float64(num1), num2)
	}

	num1, err := toFloat64(inputs[0].Value)
	if err != nil {
		return Expression{}, err
	}

	num2, err := toInt64(inputs[1].Value)
	if err != nil {
		return Expression{}, err
	}

	return decimals(num1, float64(num2))
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("value is not an integer")
	}
}

func toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("value is not a decimal")
	}
}
