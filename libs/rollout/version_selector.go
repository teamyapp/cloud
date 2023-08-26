package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
)

type VersionSelector interface {
	GetVersionNumber(viewerID uint64) (int, *errs.Error)
}

type StaticVersionSelector struct {
	versionNumber int
}

var _ VersionSelector = (*StaticVersionSelector)(nil)

func (s *StaticVersionSelector) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	return s.versionNumber, nil
}

func NewStaticVersionSelector(
	versionNumber int,
) *StaticVersionSelector {
	return &StaticVersionSelector{
		versionNumber: versionNumber,
	}
}

type ExperimentVersionSelector struct {
	store          Store
	randomGen      randgen.RandomNumberGenerator
	versionNumbers []int
}

var _ VersionSelector = (*ExperimentVersionSelector)(nil)

func (u *ExperimentVersionSelector) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	versionNumber, err := u.store.GetViewerVersionNumber(viewerID)
	if err != nil {
		return 0, err
	}

	if versionNumber != nil {
		return *versionNumber, nil
	}

	index := u.randomGen.RandomInt(len(u.versionNumbers))
	newVersionNumber := u.versionNumbers[index]
	err = u.store.SetViewerVersionNumber(viewerID, newVersionNumber)
	return newVersionNumber, err
}

func NewExperimentVersionSelector(
	store Store,
	randomGen randgen.RandomNumberGenerator,
	versionNumbers []int,
) *ExperimentVersionSelector {
	return &ExperimentVersionSelector{
		store:          store,
		randomGen:      randomGen,
		versionNumbers: versionNumbers,
	}
}
