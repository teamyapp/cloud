package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Rollout interface {
	IsActive() bool
	GetVersionNumber(viewerID uint64) (int, *errs.Error)
}
