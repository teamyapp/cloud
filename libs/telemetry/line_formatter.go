package telemetry

import (
	"encoding/json"
	"fmt"
	"strings"
)

type LineFormatter interface {
	FormatLine(props Props) (string, error)
}

type JSONLineFormatter struct {
}

var _ LineFormatter = (*JSONLineFormatter)(nil)

func (j JSONLineFormatter) FormatLine(props Props) (string, error) {
	buf, err := json.Marshal(props)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func NewJSONLineFormatter() JSONLineFormatter {
	return JSONLineFormatter{}
}

type OrderedColumnLineFormatter struct {
	columns []string
}

var _ LineFormatter = (*OrderedColumnLineFormatter)(nil)

func (o OrderedColumnLineFormatter) FormatLine(props Props) (string, error) {
	var pairs []string
	for _, column := range o.columns {
		val, ok := props[column]
		if !ok {
			continue
		}

		pairs = append(pairs, fmt.Sprintf("(%v:%+v)", column, val))
	}

	return strings.Join(pairs, " "), nil
}

func NewOrderedColumnLineFormatter(columns []string) OrderedColumnLineFormatter {
	return OrderedColumnLineFormatter{
		columns: columns,
	}
}
