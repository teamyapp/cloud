package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Static struct {
	versionNumber int
}

var _ Rollout = (*Static)(nil)

func (s *Static) IsActive() bool {
	return true
}

func (s *Static) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	return s.versionNumber, nil
}

func NewStatic(versionNumber int) Static {
	return Static{
		versionNumber: versionNumber,
	}
}
