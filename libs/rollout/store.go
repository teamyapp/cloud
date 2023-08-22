package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Store interface {
	GetViewerVersionNumber(viewerID uint64) (*int, *errs.Error)
	SetViewerVersionNumber(viewerID uint64, versionNumber int) *errs.Error
	GetTotalViewers(defaultViewers int) (int, *errs.Error)
	SetTotalViewers(totalViewers int) *errs.Error
	GetIsActivated(viewerID uint64) (*bool, *errs.Error)
	SetIsActivated(viewerID uint64, isActivated bool) *errs.Error
	GetIsRolloutEnabled(defaultIsRolloutEnabled bool) (bool, *errs.Error)
	SetIsRolloutEnabled(isRolloutEnabled bool) *errs.Error
	GetBucketIndex(defaultBucketIndex int) (int, *errs.Error)
	SetBucketIndex(bucketIndex int) *errs.Error
}
