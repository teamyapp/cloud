package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Store interface {
	GetViewerVersionNumber(rolloutID uint64, viewerID uint64) (*int, *errs.Error)
	SetViewerVersionNumber(rolloutID uint64, viewerID uint64, versionNumber int) *errs.Error
	GetTotalViewers(rolloutID uint64) (int, *errs.Error)
	SetTotalViewers(rolloutID uint64, totalViewers int) *errs.Error
}
