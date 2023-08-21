package rollout

import (
	"time"

	"github.com/teamyapp/cloud/libs/errs"
)

type TimeRange struct {
	versionNumber int
	startAt       *time.Time
	endAt         *time.Time
}

var _ Rollout = (*TimeRange)(nil)

func (t *TimeRange) IsActive() bool {
	return isInTimeRange(t.startAt, t.endAt)
}

func (t *TimeRange) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	return t.versionNumber, nil
}

func NewTimeRangeRollout(
	versionNumber int,
	startAt *time.Time,
	endAt *time.Time,
) TimeRange {
	return TimeRange{
		versionNumber: versionNumber,
		startAt:       startAt,
		endAt:         endAt,
	}
}

func isInTimeRange(startAt *time.Time, endAt *time.Time) bool {
	now := time.Now().UTC()
	if startAt != nil && now.Before(*startAt) {
		return false
	}

	if endAt != nil && now.After(*endAt) {
		return false
	}

	return true
}
