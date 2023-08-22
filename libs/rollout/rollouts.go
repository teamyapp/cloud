package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type OrderedRollouts []Rollout

func (o OrderedRollouts) GetVersionNumber(viewerID uint64) (*int, *errs.Error) {
	for index := len(o) - 1; index >= 0; index-- {
		rollout := o[index]
		isActive, err := rollout.IsActive(viewerID)
		if err != nil {
			return nil, err
		}

		if isActive {
			versionNumber, err := rollout.GetVersionNumber(viewerID)
			return &versionNumber, err
		}
	}

	return nil, nil
}
