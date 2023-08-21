package rollout

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type MaxViewers struct {
	store         Store
	id            uint64
	versionNumber int
	totalViewers  int
	maxViewers    int
}

var _ Rollout = (*MaxViewers)(nil)

func (m *MaxViewers) IsActive() bool {
	return m.totalViewers <= m.maxViewers
}

func (m *MaxViewers) GetVersionNumber(viewerID uint64) (int, *errs.Error) {
	m.totalViewers++
	err := m.store.SetTotalViewers(m.id, m.totalViewers)
	return m.versionNumber, err
}

func NewMaxViewers(store Store, id uint64, versionNumber int, maxViewers int) (MaxViewers, *errs.Error) {
	totalViewers, err := store.GetTotalViewers(id)
	if err != nil {
		return MaxViewers{}, err
	}

	return MaxViewers{
		store:         store,
		id:            id,
		versionNumber: versionNumber,
		totalViewers:  totalViewers,
		maxViewers:    maxViewers,
	}, nil
}
