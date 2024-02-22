package query

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Function func(...Expression) (Expression, error)

type Environment interface {
	Get(identifier string) (Expression, bool)
	Set(identifier string, expression Expression)
}

type MapEnvironment map[string]Expression

var _ Environment = MapEnvironment{}

func (m MapEnvironment) Get(identifier string) (Expression, bool) {
	value, ok := m[identifier]
	return value, ok
}

func (m MapEnvironment) Set(identifier string, expression Expression) {
	m[identifier] = expression
}

func Evaluate(environment Environment, expression Expression) (Expression, error) {
	switch expression.Type {
	case ValueExpressionType:
		return expression, nil
	case InvocationExpressionType:
		return evaluateInvocation(environment, expression)
	case IdentifierExpressionType:
		return evaluateIdentifier(environment, expression)
	default:
		return Expression{}, fmt.Errorf("unknown expression type: %v", expression.Type)
	}
}

func evaluateInvocation(environment Environment, expression Expression) (Expression, error) {
	funcExpression, err := Evaluate(environment, *expression.FuncExpression)
	if err != nil {
		return Expression{}, err
	}

	if funcExpression.Type != ValueExpressionType {
		return Expression{}, fmt.Errorf("expression must be value: expression=%v", funcExpression)
	}

	if funcExpression.ValueType != FunctionDataType {
		return Expression{}, fmt.Errorf("expression must be function: expression=%v", funcExpression)
	}

	inputValues := make([]Expression, 0)
	for _, inputExp := range expression.FuncInputValues {
		inputValue, err := Evaluate(environment, inputExp)
		if err != nil {
			return Expression{}, err
		}

		if inputValue.Type != ValueExpressionType {
			return Expression{}, fmt.Errorf("input expression must be value: expression=%v", inputValue)
		}

		inputValues = append(inputValues, inputValue)
	}

	funcName := funcExpression.Value.(Function)
	return funcName(inputValues...)
}

func evaluateIdentifier(environment Environment, expression Expression) (Expression, error) {
	reference := strings.Split(expression.Identifier, ".")
	valExpress, ok := environment.Get(reference[0])
	if !ok {
		return Expression{}, fmt.Errorf("unknown identifier: %v", expression.Identifier)
	}

	if valExpress.Type != ValueExpressionType {
		return Expression{}, fmt.Errorf("expression must be value: expression=%v", valExpress)
	}

	if len(reference) == 1 {
		return valExpress, nil
	}

	if valExpress.ValueType != StructDataType {
		return Expression{}, fmt.Errorf("expression must be struct: expression=%v", valExpress)
	}

	valueReflect := reflect.ValueOf(valExpress.Value)
	for _, fieldName := range reference[1:] {
		valueReflect = valueReflect.FieldByName(fieldName)
		if !valueReflect.IsValid() {
			return Expression{}, fmt.Errorf("unknown field: identifier=%v, fieldName=%v", expression.Identifier, fieldName)
		}
	}

	value := valueReflect.Interface()
	dataType := getDataType(value)
	return Expression{
		Type:      ValueExpressionType,
		ValueType: dataType,
		Value:     value,
	}, nil
}

func getDataType(value any) DataType {
	switch value.(type) {
	case int:
		return IntDataType
	case int8:
		return IntDataType
	case int16:
		return IntDataType
	case int32:
		return IntDataType
	case int64:
		return IntDataType
	case uint:
		return IntDataType
	case uint8:
		return IntDataType
	case uint16:
		return IntDataType
	case uint32:
		return IntDataType
	case uint64:
		return IntDataType
	case float32:
		return DecimalDataType
	case float64:
		return DecimalDataType
	case string:
		return StringDataType
	case bool:
		return BoolDataType
	case time.Time:
		return DatetimeDataType
	default:
		return StructDataType
	}
}
