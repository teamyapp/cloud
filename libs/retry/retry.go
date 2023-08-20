package retry

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type Retry interface {
	WithRetry(ct context.Context, execute func() *errs.Error) (int, *errs.Error)
}
