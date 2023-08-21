package rollout

import (
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
)

type ExperimentRollout struct {
	store          Store
	randomGen      randgen.RandomNumberGenerator
	id             uint64
	versionNumbers []int
	startAt        *time.Time
	endAt          *time.Time
}

var _ Rollout = (*ExperimentRollout)(nil)

func (e *ExperimentRollout) IsActive() bool {
	return isInTimeRange(e.startAt, e.endAt)
}

func (e *ExperimentRollout) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	versionNumber, err := e.store.GetViewerVersionNumber(e.id, viewerID)
	if err != nil {
		return 0, err
	}

	if versionNumber != nil {
		return *versionNumber, nil
	}

	index := e.randomGen.RandomInt(len(e.versionNumbers))
	newVersionNumber := e.versionNumbers[index]
	err = e.store.SetViewerVersionNumber(e.id, viewerID, newVersionNumber)
	return newVersionNumber, err
}

func NewExperimentRollout(
	id uint64,
	store Store,
	versionNumbers []int,
	startAt *time.Time,
	endAt *time.Time,
) ExperimentRollout {
	return ExperimentRollout{
		id:             id,
		store:          store,
		versionNumbers: versionNumbers,
		startAt:        startAt,
		endAt:          endAt,
	}
}
