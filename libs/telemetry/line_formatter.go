package telemetry

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/teamyapp/cloud/libs/errs"
)

type LineFormatter interface {
	FormatLine(props Props) (string, *errs.Error)
}

type JSONLineFormatter struct {
}

var _ LineFormatter = (*JSONLineFormatter)(nil)

func (j JSONLineFormatter) FormatLine(props Props) (string, *errs.Error) {
	buf, err := json.Marshal(props)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		return "", internalErr
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

func (o OrderedColumnLineFormatter) FormatLine(props Props) (string, *errs.Error) {
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
