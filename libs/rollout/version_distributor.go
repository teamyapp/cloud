package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
)

type VersionDistributor interface {
	GetVersionNumber(viewerID uint64) (int, *errs.Error)
}

type StaticVersionDistributor struct {
	versionNumber int
}

var _ VersionDistributor = (*StaticVersionDistributor)(nil)

func (s *StaticVersionDistributor) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	return s.versionNumber, nil
}

func NewStaticVersionDistributor(
	versionNumber int,
) *StaticVersionDistributor {
	return &StaticVersionDistributor{
		versionNumber: versionNumber,
	}
}

type ExperimentVersionDistributor struct {
	store          Store
	randomGen      randgen.RandomNumberGenerator
	activator      Activator
	versionNumbers []int
}

var _ VersionDistributor = (*ExperimentVersionDistributor)(nil)

func (u *ExperimentVersionDistributor) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
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

func NewExperimentVersionDistributor(
	store Store,
	randomGen randgen.RandomNumberGenerator,
	versionNumbers []int,
) *ExperimentVersionDistributor {
	return &ExperimentVersionDistributor{
		store:          store,
		randomGen:      randomGen,
		versionNumbers: versionNumbers,
	}
}
