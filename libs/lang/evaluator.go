package lang

import (
	"fmt"
)

type DataType string

type Value struct {
	dataType DataType
	data     any
}

func Evaluate(expression Expression) (any, *Err) {
	switch expression.Type {
	case TernaryExpressionType:
		return evaluateTernaryExpression(expression)
	case BinaryExpressionType:
		return evaluateBinaryExpression(expression)
	case UnaryExpressionType:
		return evaluateUnaryExpression(expression)
	case LiteralExpressionType:
		return evaluateLiteralExpression(expression), nil
	case GroupingExpressionType:
		return evaluateGroupingExpression(expression)
	case ExpressionListExpressionType:
		return evaluateExpressionListExpression(expression)
	}

	return nil, &Err{
		Message: fmt.Sprintf("unsupported expression type: %v", expression.Type),
		Line:    expression.Line,
		Column:  expression.Column,
	}
}

func evaluateTernaryExpression(expression Expression) (any, *Err) {
	condition, err := Evaluate(*expression.TernaryConditionExpression)
	if err != nil {
		return nil, err
	}

	if condition == true {
		return Evaluate(*expression.TernaryTrueExpression)
	} else {
		return Evaluate(*expression.TernaryFalseExpression)
	}
}

func evaluateBinaryExpression(expression Expression) (any, *Err) {
	leftValue, err := Evaluate(*expression.BinaryLeftExpression)
	if err != nil {
		return nil, err
	}

	rightValue, err := Evaluate(*expression.BinaryRightExpression)
	if err != nil {
		return nil, err
	}

	switch expression.Operator.Type {
	case LogicalEqualTokenType:
		return leftValue == rightValue, nil
	case LogicalNotEqualTokenType:
		return leftValue != rightValue, nil
	case LogicalOrTokenType:
		leftBool, err := toBoolean(leftValue, expression.Operator.Line, expression.Operator.Column)
		if err != nil {
			return nil, err
		}

		rightBool, err := toBoolean(rightValue, expression.Operator.Line, expression.Operator.Column)
		if err != nil {
			return nil, err
		}

		return leftBool || rightBool, nil
	case LogicalAndTokenType:
		leftBool, err := toBoolean(leftValue, expression.Operator.Line, expression.Operator.Column)
		if err != nil {
			return nil, err
		}

		rightBool, err := toBoolean(rightValue, expression.Operator.Line, expression.Operator.Column)
		if err != nil {
			return nil, err
		}

		return leftBool && rightBool, nil
	case GreaterThanTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 > int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 > float2, nil
			})
	case GreaterThanOrEqualTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 >= int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 >= float2, nil
			})
	case LessThanTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 < int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 < float2, nil
			})
	case LessThanOrEqualTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 <= int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 <= float2, nil
			})
	case BitwiseOrTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				return int64(byte(int1) | byte(int2)), nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float64(byte(float1) | byte(float2)), nil
			})
	case BitwiseXorTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				return int64(byte(int1) ^ byte(int2)), nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float64(byte(float1) ^ byte(float2)), nil
			})
	case BitwiseAndTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				return int64(byte(int1) & byte(int2)), nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float64(byte(float1) & byte(float2)), nil
			})
	case BitwiseLeftShiftTokenType:
		switch typedLeft := leftValue.(type) {
		case int64:
			switch typedRight := rightValue.(type) {
			case int64:
				return typedLeft << typedRight, nil
			}
		case float64:
			switch typedRight := rightValue.(type) {
			case int64:
				return float64(byte(typedLeft) << typedRight), nil
			}
		}

		return nil, &Err{
			Message: fmt.Sprintf("unsupported value type for %v: val1=%v, val2=%v",
				expression.Operator.Type,
				leftValue,
				rightValue),
			Line:   expression.Operator.Line,
			Column: expression.Operator.Column,
		}
	case BitwiseRightShiftTokenType:
		switch typedLeft := leftValue.(type) {
		case int64:
			switch typedRight := rightValue.(type) {
			case int64:
				return typedLeft >> typedRight, nil
			}
		case float64:
			switch typedRight := rightValue.(type) {
			case int64:
				return byte(typedLeft) >> typedRight, nil
			}
		}

		return nil, &Err{
			Message: fmt.Sprintf("unsupported value type for %v: val1=%v, val2=%v",
				expression.Operator.Type,
				leftValue,
				rightValue),
			Line:   expression.Operator.Line,
			Column: expression.Operator.Column,
		}
	case AddTokenType:
		typedLeft, leftOk := leftValue.(string)
		typedRight, rightOk := rightValue.(string)
		if leftOk && rightOk {
			return typedLeft + typedRight, nil
		}

		if leftOk {
			return typedLeft + fmt.Sprintf("%v", rightValue), nil
		}

		if rightOk {
			return fmt.Sprintf("%v", leftValue) + typedRight, nil
		}

		return consumeNumbers(leftValue, rightValue, expression.Line, expression.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				return int1 + int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float1 + float2, nil
			})
	case MinusTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Line, expression.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				return int1 - int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float1 - float2, nil
			})
	case StarTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Line, expression.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				return int1 * int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float1 * float2, nil
			})
	case DivideTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Operator.Line, expression.Operator.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				if int2 == 0 {
					return nil, &Err{
						Message: "division by zero",
						Line:    expression.Operator.Line,
						Column:  expression.Operator.Column,
					}
				}

				return int1 / int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				if float2 == 0 {
					return nil, &Err{
						Message: "division by zero",
						Line:    expression.Operator.Line,
						Column:  expression.Operator.Column,
					}
				}

				return float1 / float2, nil
			})
	case ModuloTokenType:
		return consumeNumbers(leftValue, rightValue, expression.Line, expression.Column,
			func(int1 int64, int2 int64) (any, *Err) {
				if int2 == 0 {
					return nil, &Err{
						Message: "division by zero",
						Line:    expression.Operator.Line,
						Column:  expression.Operator.Column,
					}
				}

				return int1 % int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return nil, &Err{
					Message: "unsupported operator %v for float: %v",
					Line:    expression.Operator.Line,
					Column:  expression.Operator.Column,
				}
			})
	}

	return nil, &Err{
		Message: fmt.Sprintf("unsupported operator %v", expression.Operator),
		Line:    expression.Operator.Line,
		Column:  expression.Operator.Column,
	}
}

func evaluateUnaryExpression(expression Expression) (any, *Err) {
	value, err := Evaluate(*expression.UnaryExpression)
	if err != nil {
		return nil, err
	}

	switch expression.Operator.Type {
	case LogicalNotTokenType:
		boolVal, err := toBoolean(value, expression.Operator.Line, expression.Operator.Column)
		if err != nil {
			return false, err
		}

		return !boolVal, nil
	case MinusTokenType:
		switch typedVal := value.(type) {
		case int64:
			return -typedVal, nil
		case float64:
			return -typedVal, nil
		}

		return nil, &Err{
			Message: fmt.Sprintf("unsupported data type for %v: %v", MinusTokenType, value),
			Line:    expression.Operator.Line,
			Column:  expression.Operator.Column,
		}
	case BitwiseNotTokenType:
		switch typedVal := value.(type) {
		case int64:
			return int64(^byte(typedVal)), nil
		case float64:
			return float64(^byte(typedVal)), nil
		}

		return nil, &Err{
			Message: fmt.Sprintf("unsupported data type for %v: %v", BitwiseNotTokenType, value),
			Line:    expression.Operator.Line,
			Column:  expression.Operator.Column,
		}
	}

	return nil, &Err{
		Message: fmt.Sprintf("unsupported operator %v", expression.Operator),
		Line:    expression.Operator.Line,
		Column:  expression.Operator.Column,
	}
}

func consumeNumbers[Result any](
	val1 any,
	val2 any,
	line int,
	column int,
	compareInts func(int1 int64, int2 int64) (Result, *Err),
	compareFloats func(float1 float64, float2 float64) (Result, *Err),
) (Result, *Err) {
	switch typedVal1 := val1.(type) {
	case int64:
		switch typedVal2 := val2.(type) {
		case int64:
			return compareInts(typedVal1, typedVal2)
		}
	case float64:
		switch typedVal2 := val2.(type) {
		case float64:
			return compareFloats(typedVal1, typedVal2)
		}
	}

	return *new(Result), &Err{
		Message: fmt.Sprintf("unsupported value type: val1=%v, val2=%v", val1, val2),
		Line:    line,
		Column:  column,
	}
}

func toBoolean(value any, line int, column int) (bool, *Err) {
	switch typedVal := value.(type) {
	case bool:
		return typedVal, nil
	case int64:
		return value != 0, nil
	case float64:
		return value != 0, nil
	}

	return false, &Err{
		Message: fmt.Sprintf("cannot convert value to boolean: %v", value),
		Line:    line,
		Column:  column,
	}
}

func evaluateLiteralExpression(expression Expression) any {
	return expression.Literal.Value
}

func evaluateGroupingExpression(expression Expression) (any, *Err) {
	return Evaluate(*expression.GroupInnerExpression)
}

func evaluateExpressionListExpression(expression Expression) (any, *Err) {
	var result any
	for _, expr := range expression.ExpressionList {
		var err *Err
		result, err = Evaluate(expr)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
