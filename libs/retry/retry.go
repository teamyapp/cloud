package retry

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Retry interface {
	WithRetry(execute func() *errs.Error) (int, *errs.Error)
}
