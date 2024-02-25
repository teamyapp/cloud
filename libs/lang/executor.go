package lang

import (
	"fmt"
	"io"
)

type DataType string

type Value struct {
	dataType DataType
	data     any
}

type Executor struct {
	environment *Environment
	output      io.Writer
}

func (e *Executor) Execute(statements []Statement) ([]any, *Err) {
	values, signal, err := e.executeStatements(statements)
	if err != nil {
		return values, err
	}

	if signal != nil {
		switch signal.Type {
		case BreakSignalType:
			return values, &Err{
				Message:           "break must appear inside loop",
				Line:              signal.Line,
				Column:            signal.Column,
				FromGeneratedCode: signal.IsGenerated,
			}
		case ContinueSignalType:
			return values, &Err{
				Message:           "continue must appear inside loop",
				Line:              signal.Line,
				Column:            signal.Column,
				FromGeneratedCode: signal.IsGenerated,
			}
		}
	}

	return values, nil
}

func (e *Executor) executeStatements(statements []Statement) ([]any, *Signal, *Err) {
	var values []any
	for _, statement := range statements {
		value, hasValue, signal, err := e.executeSingleStatement(statement)
		if err != nil {
			return values, nil, err
		}

		if signal != nil {
			return values, signal, nil
		}

		if hasValue {
			values = append(values, value)
		}
	}

	return values, nil, nil
}

func (e *Executor) executeSingleStatement(statement Statement) (any, bool, *Signal, *Err) {
	switch statement.Type {
	case PrintStatementType:
		return nil, false, nil, e.executePrintStatement(statement)
	case ExpressionStatementType:
		value, err := e.evaluateExpressionStatement(statement)
		if err != nil {
			return nil, false, nil, err
		}

		return value, true, nil, nil
	case LetStatementType:
		return nil, false, nil, e.executeLetStatement(statement)
	case BlockStatementType:
		signal, err := e.executeBlockStatement(statement)
		return nil, false, signal, err
	case IfStatementType:
		signal, err := e.executeIfStatement(statement)
		return nil, false, signal, err
	case WhileStatementType:
		return nil, false, nil, e.executeWhileStatement(statement)
	case BreakStatementType:
		return nil, false, &Signal{
			Type:        BreakSignalType,
			Line:        statement.Line,
			Column:      statement.Column,
			IsGenerated: statement.IsGenerated,
		}, nil
	case ContinueStatementType:
		return nil, false, &Signal{
			Type:        ContinueSignalType,
			Line:        statement.Line,
			Column:      statement.Column,
			IsGenerated: statement.IsGenerated,
		}, nil
	}

	return nil, false, nil, &Err{
		Message:           fmt.Sprintf("unknown statement type: %v", statement.Type),
		Line:              statement.Line,
		Column:            statement.Column,
		FromGeneratedCode: statement.IsGenerated,
	}
}

func (e *Executor) executeWhileStatement(statement Statement) *Err {
	for {
		conditionVal, err := e.evaluateExpression(*statement.WhileConditionExpression)
		if err != nil {
			return err
		}

		conditionBool, err := toBoolean(
			conditionVal,
			statement.WhileConditionExpression.Line,
			statement.WhileConditionExpression.Column,
			statement.WhileConditionExpression.IsGenerated)
		if err != nil {
			return err
		}

		if !conditionBool {
			break
		}
		_, _, signal, err := e.executeSingleStatement(*statement.WhileBodyStatement)
		if err != nil {
			return err
		}

		if signal != nil {
			switch signal.Type {
			case BreakSignalType:
				return nil
			case ContinueSignalType:
				continue
			}
		}
	}

	return nil
}

func (e *Executor) executeIfStatement(statement Statement) (*Signal, *Err) {
	value, err := e.evaluateExpression(*statement.IfConditionExpression)
	if err != nil {
		return nil, err
	}

	valueBool, err := toBoolean(
		value,
		statement.IfConditionExpression.Line,
		statement.IfConditionExpression.Column,
		statement.IfConditionExpression.IsGenerated)
	if err != nil {
		return nil, err
	}

	if valueBool {
		_, _, signal, err := e.executeSingleStatement(*statement.IfTrueBranchStatement)
		return signal, err
	}

	if statement.IfFalseBranchStatement != nil {
		_, _, signal, err := e.executeSingleStatement(*statement.IfFalseBranchStatement)
		return signal, err
	}

	return nil, nil
}

func (e *Executor) executeBlockStatement(statement Statement) (*Signal, *Err) {
	prevEnvironment := e.environment
	e.environment = e.environment.NewInnerEnvironment()
	_, signal, err := e.executeStatements(statement.BlockInnerStatements)
	e.environment = prevEnvironment
	return signal, err
}

func (e *Executor) executeLetStatement(statement Statement) *Err {
	var value any
	if statement.LetInitializerExpression != nil {
		var err *Err
		value, err = e.evaluateExpression(*statement.LetInitializerExpression)
		if err != nil {
			return err
		}

		e.environment.DefineWithInitializer(*statement.LetIdentifier, value)
		return nil
	}

	e.environment.Define(*statement.LetIdentifier)
	return nil
}

func (e *Executor) executePrintStatement(statement Statement) *Err {
	value, err := e.evaluateExpression(*statement.PrintArgExpression)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprint(e.output, toString(value))
	return nil
}

func (e *Executor) evaluateExpressionStatement(statement Statement) (any, *Err) {
	return e.evaluateExpression(*statement.StatementExpression)
}

func (e *Executor) evaluateExpression(expression Expression) (any, *Err) {
	switch expression.Type {
	case TernaryExpressionType:
		return e.evaluateTernaryExpression(expression)
	case BinaryExpressionType:
		return e.evaluateBinaryExpression(expression)
	case UnaryExpressionType:
		return e.evaluateUnaryExpression(expression)
	case LiteralExpressionType:
		return evaluateLiteralExpression(expression), nil
	case GroupingExpressionType:
		return e.evaluateGroupingExpression(expression)
	case ExpressionListExpressionType:
		return e.evaluateExpressionListExpression(expression)
	case IdentifierExpressionType:
		return e.evaluateIdentifierExpression(expression)
	case AssignmentExpressionType:
		return e.evaluateAssignmentExpression(expression)
	}

	return nil, &Err{
		Message:           fmt.Sprintf("unsupported expression type: %v", expression.Type),
		Line:              expression.Line,
		Column:            expression.Column,
		FromGeneratedCode: expression.IsGenerated,
	}
}

func (e *Executor) evaluateAssignmentExpression(expression Expression) (any, *Err) {
	value, err := e.evaluateExpression(*expression.AssignmentValueExpression)
	if err != nil {
		return nil, err
	}

	err = e.environment.Assign(expression.Identifier, value)
	return value, err
}

func (e *Executor) evaluateIdentifierExpression(expression Expression) (any, *Err) {
	return e.environment.Get(expression.Identifier)
}

func (e *Executor) evaluateTernaryExpression(expression Expression) (any, *Err) {
	conditionExpr, err := e.evaluateExpression(*expression.TernaryConditionExpression)
	if err != nil {
		return nil, err
	}

	condition, err := toBoolean(
		conditionExpr,
		expression.TernaryConditionExpression.Line,
		expression.TernaryConditionExpression.Column,
		expression.TernaryConditionExpression.IsGenerated)
	if err != nil {
		return nil, err
	}

	if condition == true {
		return e.evaluateExpression(*expression.TernaryTrueExpression)
	} else {
		return e.evaluateExpression(*expression.TernaryFalseExpression)
	}
}

func (e *Executor) evaluateBinaryExpression(expression Expression) (any, *Err) {
	leftValue, err := e.evaluateExpression(*expression.BinaryLeftExpression)
	if err != nil {
		return nil, err
	}

	switch expression.Operator.Type {
	case LogicalEqualTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return leftValue == rightValue, nil
	case LogicalNotEqualTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return leftValue != rightValue, nil
	case LogicalOrTokenType:
		leftBool, err := toBoolean(
			leftValue,
			expression.BinaryLeftExpression.Line,
			expression.BinaryLeftExpression.Column,
			expression.BinaryLeftExpression.IsGenerated)
		if err != nil {
			return nil, err
		}

		if leftBool {
			return true, nil
		}

		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		rightBool, err := toBoolean(
			rightValue,
			expression.BinaryRightExpression.Line,
			expression.BinaryRightExpression.Column,
			expression.BinaryRightExpression.IsGenerated)
		if err != nil {
			return nil, err
		}

		return leftBool || rightBool, nil
	case LogicalAndTokenType:
		leftBool, err := toBoolean(
			leftValue,
			expression.BinaryLeftExpression.Line,
			expression.BinaryLeftExpression.Column,
			expression.BinaryLeftExpression.IsGenerated)
		if err != nil {
			return nil, err
		}

		if !leftBool {
			return false, nil
		}

		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		rightBool, err := toBoolean(
			rightValue,
			expression.BinaryRightExpression.Line,
			expression.BinaryRightExpression.Column,
			expression.BinaryRightExpression.IsGenerated)
		if err != nil {
			return nil, err
		}

		return leftBool && rightBool, nil
	case GreaterThanTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 > int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 > float2, nil
			})
	case GreaterThanOrEqualTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 >= int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 >= float2, nil
			})
	case LessThanTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 < int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 < float2, nil
			})
	case LessThanOrEqualTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (bool, *Err) {
				return int1 <= int2, nil
			},
			func(float1 float64, float2 float64) (bool, *Err) {
				return float1 <= float2, nil
			})
	case BitwiseOrTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				return int64(byte(int1) | byte(int2)), nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float64(byte(float1) | byte(float2)), nil
			})
	case BitwiseXorTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				return int64(byte(int1) ^ byte(int2)), nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float64(byte(float1) ^ byte(float2)), nil
			})
	case BitwiseAndTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				return int64(byte(int1) & byte(int2)), nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float64(byte(float1) & byte(float2)), nil
			})
	case BitwiseLeftShiftTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

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
			Line:              expression.Operator.Line,
			Column:            expression.Operator.Column,
			FromGeneratedCode: expression.Operator.IsGenerated,
		}
	case BitwiseRightShiftTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

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
			Line:              expression.Operator.Line,
			Column:            expression.Operator.Column,
			FromGeneratedCode: expression.Operator.IsGenerated,
		}
	case AddTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

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

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				return int1 + int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float1 + float2, nil
			})
	case MinusTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				return int1 - int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float1 - float2, nil
			})
	case StarTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				return int1 * int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return float1 * float2, nil
			})
	case DivideTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				if int2 == 0 {
					return nil, &Err{
						Message:           "division by zero",
						Line:              expression.Operator.Line,
						Column:            expression.Operator.Column,
						FromGeneratedCode: expression.Operator.IsGenerated,
					}
				}

				return int1 / int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				if float2 == 0 {
					return nil, &Err{
						Message:           "division by zero",
						Line:              expression.Operator.Line,
						Column:            expression.Operator.Column,
						FromGeneratedCode: expression.Operator.IsGenerated,
					}
				}

				return float1 / float2, nil
			})
	case ModuloTokenType:
		rightValue, err := e.evaluateExpression(*expression.BinaryRightExpression)
		if err != nil {
			return nil, err
		}

		return consumeNumbers(
			leftValue,
			rightValue,
			expression.Operator.Line,
			expression.Operator.Column,
			expression.Operator.IsGenerated,
			func(int1 int64, int2 int64) (any, *Err) {
				if int2 == 0 {
					return nil, &Err{
						Message:           "division by zero",
						Line:              expression.Operator.Line,
						Column:            expression.Operator.Column,
						FromGeneratedCode: expression.Operator.IsGenerated,
					}
				}

				return int1 % int2, nil
			},
			func(float1 float64, float2 float64) (any, *Err) {
				return nil, &Err{
					Message:           "unsupported operator %v for float: %v",
					Line:              expression.Operator.Line,
					Column:            expression.Operator.Column,
					FromGeneratedCode: expression.Operator.IsGenerated,
				}
			})
	}

	return nil, &Err{
		Message:           fmt.Sprintf("unsupported operator %v", expression.Operator),
		Line:              expression.Operator.Line,
		Column:            expression.Operator.Column,
		FromGeneratedCode: expression.Operator.IsGenerated,
	}
}

func (e *Executor) evaluateUnaryExpression(expression Expression) (any, *Err) {
	value, err := e.evaluateExpression(*expression.UnaryExpression)
	if err != nil {
		return nil, err
	}

	switch expression.Operator.Type {
	case LogicalNotTokenType:
		boolVal, err := toBoolean(
			value,
			expression.UnaryExpression.Line,
			expression.UnaryExpression.Column,
			expression.UnaryExpression.IsGenerated)
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
			Message:           fmt.Sprintf("unsupported data type for %v: %v", MinusTokenType, value),
			Line:              expression.Operator.Line,
			Column:            expression.Operator.Column,
			FromGeneratedCode: expression.Operator.IsGenerated,
		}
	case BitwiseNotTokenType:
		switch typedVal := value.(type) {
		case int64:
			return int64(^byte(typedVal)), nil
		case float64:
			return float64(^byte(typedVal)), nil
		}

		return nil, &Err{
			Message:           fmt.Sprintf("unsupported data type for %v: %v", BitwiseNotTokenType, value),
			Line:              expression.Operator.Line,
			Column:            expression.Operator.Column,
			FromGeneratedCode: expression.Operator.IsGenerated,
		}
	}

	return nil, &Err{
		Message:           fmt.Sprintf("unsupported operator %v", expression.Operator),
		Line:              expression.Operator.Line,
		Column:            expression.Operator.Column,
		FromGeneratedCode: expression.Operator.IsGenerated,
	}
}

func consumeNumbers[Result any](
	val1 any,
	val2 any,
	line int,
	column int,
	isGenerated bool,
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
		Message:           fmt.Sprintf("unsupported value type: val1=%v, val2=%v", val1, val2),
		Line:              line,
		Column:            column,
		FromGeneratedCode: isGenerated,
	}
}

func toBoolean(value any, line int, column int, isGenerated bool) (bool, *Err) {
	switch typedVal := value.(type) {
	case bool:
		return typedVal, nil
	case int64:
		return value != 0, nil
	case float64:
		return value != 0, nil
	}

	return false, &Err{
		Message:           fmt.Sprintf("cannot convert value to boolean: %v", value),
		Line:              line,
		Column:            column,
		FromGeneratedCode: isGenerated,
	}
}

func evaluateLiteralExpression(expression Expression) any {
	return expression.Literal.Value
}

func (e *Executor) evaluateGroupingExpression(expression Expression) (any, *Err) {
	return e.evaluateExpression(*expression.GroupInnerExpression)
}

func (e *Executor) evaluateExpressionListExpression(expression Expression) (any, *Err) {
	var result any
	for _, expr := range expression.ExpressionList {
		var err *Err
		result, err = e.evaluateExpression(expr)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func toString(value any) string {
	if value == nil {
		return "nil"
	}

	return fmt.Sprintf("%v", value)
}

func NewExecutor(
	environment *Environment,
	output io.Writer,
) *Executor {
	return &Executor{
		environment: environment,
		output:      output,
	}
}
