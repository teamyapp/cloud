package rollout

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type Store interface {
	GetIsRolloutEnabled(ct context.Context, defaultIsRolloutEnabled bool) (bool, *errs.Error)
	SetIsRolloutEnabled(ct context.Context, isRolloutEnabled bool) *errs.Error
}

type Rollout struct {
	store           Store
	isEnabled       bool
	activator       Activator
	versionSelector VersionSelector
}

func (r *Rollout) IsActive(ct context.Context, viewerID uint64) (bool, *errs.Error) {
	if !r.isEnabled {
		return false, nil
	}

	return r.activator.IsActive(ct, viewerID)
}

func (r *Rollout) GetVersionNumber(ct context.Context, viewerID uint64) (int, *errs.Error) {
	return r.versionSelector.GetVersionNumber(ct, viewerID)
}

func (r *Rollout) SetIsEnabled(ct context.Context, isEnabled bool) *errs.Error {
	err := r.store.SetIsRolloutEnabled(ct, isEnabled)
	if err != nil {
		return err
	}

	r.isEnabled = isEnabled
	return nil
}

func NewRollout(
	ct context.Context,
	store Store,
	activator Activator,
	versionSelector VersionSelector,
) (Rollout, *errs.Error) {
	isEnabled, err := store.GetIsRolloutEnabled(ct, true)
	if err != nil {
		return Rollout{}, err
	}

	return Rollout{
		isEnabled:       isEnabled,
		activator:       activator,
		versionSelector: versionSelector,
	}, nil
}
