package lang

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Result[Value any] struct {
	Value Value
	Error *errs.Error
}
