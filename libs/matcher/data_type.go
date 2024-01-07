package matcher

import (
	"errors"
	"strconv"
	"time"
)

type DataType string

const (
	IntDataType              DataType = "int"
	DecimalDataType          DataType = "decimal"
	BoolDataType             DataType = "bool"
	StringDataType           DataType = "string"
	RuneDataType             DataType = "rune"
	DatetimeDataType         DataType = "datetime"
	FilterExpressionDataType DataType = "filterExpression"
)

func ParseValue(dataType DataType, input string) (interface{}, error) {

	switch dataType {
	case IntDataType:
		return strconv.ParseInt(input, 10, 64)
	case DecimalDataType:
		return strconv.ParseFloat(input, 64)
	case BoolDataType:
		return strconv.ParseBool(input)
	case StringDataType:
		return input, nil
	case RuneDataType:
		if len(input) != 1 {
			return nil, errors.New("must contain 1 rune")
		}

		return input[0], nil
	case DatetimeDataType:
		return time.Parse(time.RFC3339, input)
	default:
		return nil, errors.New("unknown expression")
	}
}
