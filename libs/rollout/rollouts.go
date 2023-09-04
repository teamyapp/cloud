package rollout

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type OrderedRollouts []Rollout

func (o OrderedRollouts) GetVersionNumber(ct context.Context, viewerID uint64) (*int, *errs.Error) {
	for index := len(o) - 1; index >= 0; index-- {
		rollout := o[index]
		isActive, err := rollout.IsActive(ct, viewerID)
		if err != nil {
			return nil, err
		}

		if isActive {
			versionNumber, err := rollout.GetVersionNumber(ct, viewerID)
			return &versionNumber, err
		}
	}

	return nil, nil
}
