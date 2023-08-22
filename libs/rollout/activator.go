package rollout

import (
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
)

type Activator interface {
	IsActive(viewerID uint64) (bool, *errs.Error)
}

type StaticActivator struct {
}

var _ Activator = (*StaticActivator)(nil)

func (s StaticActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	return true, nil
}

func NewStaticActivator() StaticActivator {
	return StaticActivator{}
}

type TimeRangeActivator struct {
	startAt *time.Time
	endAt   *time.Time
}

var _ Activator = (*TimeRangeActivator)(nil)

func (t *TimeRangeActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	now := time.Now().UTC()
	if t.startAt != nil && now.Before(*t.startAt) {
		return false, nil
	}

	if t.endAt != nil && now.After(*t.endAt) {
		return false, nil
	}

	return true, nil
}

func NewTimeRangeActivator(startAt *time.Time, endAt *time.Time) TimeRangeActivator {
	return TimeRangeActivator{
		startAt: startAt,
		endAt:   endAt,
	}
}

type MaxViewersActivator struct {
	store        Store
	totalViewers int
	maxViewers   int
}

var _ Activator = (*MaxViewersActivator)(nil)

func (m *MaxViewersActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	isActivated, err := m.store.GetIsActivated(viewerID)
	if err != nil {
		return false, err
	}

	if isActivated != nil {
		return *isActivated, nil
	}

	if m.totalViewers >= m.maxViewers {
		return false, nil
	}

	m.totalViewers++
	err = m.store.SetTotalViewers(m.totalViewers)
	if err != nil {
		return false, err
	}

	err = m.store.SetIsActivated(viewerID, true)
	return true, err
}

func NewMaxViewersActivator(
	store Store,
	maxViewers int,
) (MaxViewersActivator, *errs.Error) {
	totalViewers, err := store.GetTotalViewers(0)
	if err != nil {
		return MaxViewersActivator{}, err
	}

	return MaxViewersActivator{
		store:        store,
		totalViewers: totalViewers,
		maxViewers:   maxViewers,
	}, nil
}

type PercentageActivator struct {
	store      Store
	randomGen  randgen.RandomNumberGenerator
	percentage int
}

var _ Activator = (*PercentageActivator)(nil)

func (p *PercentageActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	isActivated, err := p.store.GetIsActivated(viewerID)
	if err != nil {
		return false, err
	}

	if isActivated != nil {
		return *isActivated, nil
	}

	isActive := p.randomGen.RandomInt(100) < p.percentage
	err = p.store.SetIsActivated(viewerID, isActive)
	return isActive, err
}

func NewPercentageActivator(
	randomGen randgen.RandomNumberGenerator,
	percentage int,
) PercentageActivator {
	return PercentageActivator{
		randomGen:  randomGen,
		percentage: percentage,
	}
}

type Bucket struct {
	Percentage      int
	MinimalBakeTime time.Duration
}

type IncrementalActivator struct {
	store         Store
	randomGen     randgen.RandomNumberGenerator
	buckets       []Bucket
	bucketIndex   int
	bucketStartAt time.Time
}

var _ Activator = (*IncrementalActivator)(nil)

func (i *IncrementalActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	isActivated, err := i.store.GetIsActivated(viewerID)
	if err != nil {
		return false, err
	}

	if isActivated != nil {
		return *isActivated, nil
	}

	if i.bucketIndex >= len(i.buckets) {
		return true, nil
	}

	now := time.Now().UTC()
	timeElapsed := now.Sub(i.bucketStartAt)
	if timeElapsed >= i.buckets[i.bucketIndex].MinimalBakeTime {
		i.bucketIndex++
		i.bucketStartAt = now
		err = i.store.SetBucketIndex(i.bucketIndex + 1)
		if err != nil {
			return false, err
		}

		if i.bucketIndex >= len(i.buckets) {
			return true, nil
		}
	}

	isActive := i.randomGen.RandomInt(100) < i.buckets[i.bucketIndex].Percentage
	err = i.store.SetIsActivated(viewerID, isActive)
	return isActive, err
}

func NewIncrementalActivator(
	store Store,
	randomGen randgen.RandomNumberGenerator,
	buckets []Bucket,
) (IncrementalActivator, *errs.Error) {
	bucketIndex, err := store.GetBucketIndex(0)
	if err != nil {
		return IncrementalActivator{}, err
	}

	return IncrementalActivator{
		store:         store,
		randomGen:     randomGen,
		buckets:       buckets,
		bucketIndex:   bucketIndex,
		bucketStartAt: time.Now().UTC(),
	}, nil
}
