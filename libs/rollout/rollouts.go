package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type OrderedRollouts []Rollout

func (o OrderedRollouts) GetVersionNumber(viewerID uint64) (*int, *errs.Error) {
	for index := len(o) - 1; index >= 0; index-- {
		rollout := o[index]
		if rollout.IsActive() {
			versionNumber, err := rollout.GetVersionNumber(viewerID)
			return &versionNumber, err
		}
	}

	return nil, nil
}
