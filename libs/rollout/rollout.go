package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Rollout struct {
	store           Store
	isEnabled       bool
	activator       Activator
	versionSelector VersionSelector
}

func (r *Rollout) IsActive(viewerID uint64) (bool, *errs.Error) {
	if !r.isEnabled {
		return false, nil
	}

	return r.activator.IsActive(viewerID)
}

func (r *Rollout) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	return r.versionSelector.GetVersionNumber(viewerID)
}

func (r *Rollout) SetIsEnabled(isEnabled bool) *errs.Error {
	err := r.store.SetIsRolloutEnabled(isEnabled)
	if err != nil {
		return err
	}

	r.isEnabled = isEnabled
	return nil
}

func NewRollout(
	store Store,
	activator Activator,
	versionSelector VersionSelector,
) (Rollout, *errs.Error) {
	isEnabled, err := store.GetIsRolloutEnabled(true)
	if err != nil {
		return Rollout{}, err
	}

	return Rollout{
		isEnabled:       isEnabled,
		activator:       activator,
		versionSelector: versionSelector,
	}, nil
}
