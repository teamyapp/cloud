package matcher

import (
	"errors"
	"fmt"
	"time"
)

type Evaluator struct {
	fieldsRetriever FieldsRetriever
}

type SelectorCreator[Item any] func(attribute string) (Selector[Item], error)

func (e *Evaluator) Evaluate(expression Expression, object interface{}) bool {
	panic("implement")
}

func evaluateComparison[Item any](
	createAttributeSelector SelectorCreator[Item],
	operator Operator,
	attribute Expression,
	target Expression,
) (Filter[Item], DataType, error) {
	attributeResult, dataType, err := evaluateExpression(createAttributeSelector, attribute)
	if err != nil {
		return nil, "", err
	}
	if dataType != StringDataType {
		return nil, "", errors.New("only accept string as the 1st parameter")
	}

	targetResult, dataType, err := evaluateExpression(createAttributeSelector, target)
	if err != nil {
		return nil, "", err
	}

	switch dataType {
	case IntDataType:
		return createComparisonFilter[Item](createAttributeSelector, operator, attributeResult.(string), targetResult.(int))
	case DecimalDataType:
		return createComparisonFilter[Item](createAttributeSelector, operator, attributeResult.(string), targetResult.(float64))
	case StringDataType:
		return createComparisonFilter[Item](createAttributeSelector, operator, attributeResult.(string), targetResult.(string))
	case RuneDataType:
		return createComparisonFilter[Item](createAttributeSelector, operator, attributeResult.(string), targetResult.(rune))
	default:
		return nil, "", fmt.Errorf("unsupported data type: %v", dataType)
	}
}

func createComparisonFilter[Item any, Value Comparable](
	createAttributeSelector SelectorCreator[Item],
	operator Operator,
	attribute string,
	target Value,
) (Filter[Item], DataType, error) {
	selector, err := createAttributeSelector(attribute)
	if err != nil {
		return nil, "", err
	}

	switch operator {
	case LessThanOperator:
		return LessThan[Item](selector, target), FilterExpressionDataType, nil
	case LessThanOrEqualToOperator:
		return LessThanOrEqualTo[Item](selector, target), FilterExpressionDataType, nil
	case GreaterThanOperator:
		return GreaterThan[Item](selector, target), FilterExpressionDataType, nil
	case GreaterThanOrEqualToOperator:
		return GreaterThanOrEqualTo[Item](selector, target), FilterExpressionDataType, nil
	default:
		return nil, "", fmt.Errorf("unsupported operator: %v", operator)
	}
}

func evaluateEqualTo[Item any](
	createAttributeSelector SelectorCreator[Item],
	attribute Expression,
	target Expression,
) (Filter[Item], DataType, error) {
	attributeResult, dataType, err := evaluateExpression(createAttributeSelector, attribute)
	if err != nil {
		return nil, "", err
	}
	if dataType != StringDataType {
		return nil, "", errors.New("only accept string as the 1st parameter")
	}

	targetResult, dataType, err := evaluateExpression(createAttributeSelector, target)
	if err != nil {
		return nil, "", err
	}

	selector, err := createAttributeSelector(attributeResult.(string))
	if err != nil {
		return nil, "", err
	}

	switch dataType {
	case IntDataType:
		return EqualTo[Item, int](selector, targetResult.(int)), FilterExpressionDataType, nil
	case DecimalDataType:
		return EqualTo[Item, float64](selector, targetResult.(float64)), FilterExpressionDataType, nil
	case StringDataType:
		return EqualTo[Item, string](selector, targetResult.(string)), FilterExpressionDataType, nil
	case RuneDataType:
		return EqualTo[Item, rune](selector, targetResult.(rune)), FilterExpressionDataType, nil
	case BoolDataType:
		return EqualTo[Item, bool](selector, targetResult.(bool)), FilterExpressionDataType, nil
	case DatetimeDataType:
		return EqualTo[Item, time.Time](selector, targetResult.(time.Time)), FilterExpressionDataType, nil
	default:
		return nil, "", fmt.Errorf("unsupported data type: %v", dataType)
	}
}

func evaluateAnd[Item any](createAttributeSelector SelectorCreator[Item], filter1 Expression, filter2 Expression) (Filter[Item], DataType, error) {
	filter1Result, dataType, err := evaluateExpression(createAttributeSelector, filter1)
	if err != nil {
		return nil, "", err
	}
	if dataType != FilterExpressionDataType {
		return nil, "", errors.New("only accept filter as the 1st parameter")
	}

	filter2Result, dataType, err := evaluateExpression(createAttributeSelector, filter2)
	if err != nil {
		return nil, "", err
	}
	if dataType != FilterExpressionDataType {
		return nil, "", errors.New("only accept filter as the 2nd parameter")
	}

	return And(filter1Result.(Filter[Item]), filter2Result.(Filter[Item])), FilterExpressionDataType, nil
}

func evaluateOr[Item any](
	createAttributeSelector SelectorCreator[Item],
	filter1 Expression,
	filter2 Expression,
) (Filter[Item], DataType, error) {
	filter1Result, dataType, err := evaluateExpression(createAttributeSelector, filter1)
	if err != nil {
		return nil, "", err
	}
	if dataType != FilterExpressionDataType {
		return nil, "", errors.New("only accept filter as the 1st parameter")
	}

	filter2Result, dataType, err := evaluateExpression(createAttributeSelector, filter2)
	if err != nil {
		return nil, "", err
	}
	if dataType != FilterExpressionDataType {
		return nil, "", errors.New("only accept filter as the 2nd parameter")
	}

	return Or[Item](filter1Result.(Filter[Item]), filter2Result.(Filter[Item])), FilterExpressionDataType, nil
}

func evaluateNot[Item any](
	createAttributeSelector SelectorCreator[Item],
	filter Expression,
) (Filter[Item], DataType, error) {
	filterResult, dataType, err := evaluateExpression(createAttributeSelector, filter)
	if err != nil {
		return nil, "", err
	}
	if dataType != FilterExpressionDataType {
		return nil, "", errors.New("only accept filter as parameter")
	}

	return Not[Item](filterResult.(Filter[Item])), FilterExpressionDataType, nil
}

func evaluateContains[Item any](
	createAttributeSelector SelectorCreator[Item],
	attribute Expression,
	target Expression,
) (Filter[Item], DataType, error) {
	attributeResult, dataType, err := evaluateExpression(createAttributeSelector, attribute)
	if err != nil {
		return nil, "", err
	}
	if dataType != StringDataType {
		return nil, "", errors.New("only accept string as the 1st parameter")
	}

	targetResult, dataType, err := evaluateExpression(createAttributeSelector, target)
	if err != nil {
		return nil, "", err
	}
	if dataType != StringDataType {
		return nil, "", errors.New("only accept string as the 2nd parameter")
	}

	selector, err := createAttributeSelector(attributeResult.(string))
	if err != nil {
		return nil, "", err
	}
	return Contains[Item](selector, targetResult.(string)), FilterExpressionDataType, nil
}

func evaluateExpression[Item any](
	createAttributeSelector SelectorCreator[Item],
	expression Expression,
) (interface{}, DataType, error) {
	if expression.IsValue {
		value, err := ParseValue(expression.OutputDataType, expression.Value)
		return value, expression.OutputDataType, err
	}

	switch expression.Operator {
	case AndOperator:
		if len(expression.Inputs) != 2 {
			return nil, "", errors.New("and must have 2 parameters")
		}

		return evaluateAnd(createAttributeSelector, expression.Inputs[0], expression.Inputs[1])
	case OrOperator:
		if len(expression.Inputs) != 2 {
			return nil, "", errors.New("and must have 2 parameters")
		}

		return evaluateOr(createAttributeSelector, expression.Inputs[0], expression.Inputs[1])
	case NotOperator:
		if len(expression.Inputs) != 1 {
			return nil, "", errors.New("and must have 1 parameter")
		}

		return evaluateNot(createAttributeSelector, expression.Inputs[0])
	case EqualToOperator:
		if len(expression.Inputs) != 2 {
			return nil, "", errors.New("and must have 2 parameters")
		}

		return evaluateEqualTo(createAttributeSelector, expression.Inputs[0], expression.Inputs[1])
	case ContainsOperator:
		if len(expression.Inputs) != 2 {
			return nil, "", errors.New("and must have 2 parameters")
		}

		return evaluateContains(createAttributeSelector, expression.Inputs[0], expression.Inputs[1])
	case
		LessThanOperator,
		LessThanOrEqualToOperator,
		GreaterThanOperator,
		GreaterThanOrEqualToOperator:
		if len(expression.Inputs) != 2 {
			return nil, "", errors.New("and must have 2 parameters")
		}

		return evaluateComparison(
			createAttributeSelector,
			expression.Operator,
			expression.Inputs[0],
			expression.Inputs[1])
	default:
		return nil, "", fmt.Errorf("unknown operator: %v", expression.Operator)
	}
}

type Entity struct {
	IDAttribute     uint64
	SchemaAttribute string
}

func CreateEntityAttributeSelector(attribute string) (Selector[Entity], error) {
	switch attribute {
	case "IDAttribute":
		return func(entity Entity) interface{} {
			return entity.IDAttribute
		}, nil
	// case "SchemaAttribute":
	// 	return func(entity Entity) interface{} {
	// 		return entity.SchemaAttribute
	// 	}, nil
	default:
		return func(entity Entity) interface{} {
			return nil
		}, nil
	}
}

// create another entity attributes selector with reflect

func NewEvaluator(customFieldsRetriever map[string]CustomFieldRetriever) *Evaluator {
	fieldsRetriever := NewFieldsRetriever(customFieldsRetriever)

	return &Evaluator{
		fieldsRetriever: *fieldsRetriever,
	}
}
