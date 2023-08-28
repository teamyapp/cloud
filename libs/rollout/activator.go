package rollout

import (
	"time"

	"github.com/benbjohnson/clock"
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
	clock   clock.Clock
	startAt *time.Time
	endAt   *time.Time
}

var _ Activator = (*TimeRangeActivator)(nil)

func (t *TimeRangeActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	now := t.clock.Now().UTC()
	if t.startAt != nil && now.Before(*t.startAt) {
		return false, nil
	}

	if t.endAt != nil && now.After(*t.endAt) {
		return false, nil
	}

	return true, nil
}

func NewTimeRangeActivator(clock clock.Clock, startAt *time.Time, endAt *time.Time) *TimeRangeActivator {
	return &TimeRangeActivator{
		clock:   clock,
		startAt: startAt,
		endAt:   endAt,
	}
}

type MaxViewersActivatorStore interface {
	GetIsActivated(viewerID uint64) (*bool, *errs.Error)
	SetIsActivated(viewerID uint64, isActivated bool) *errs.Error
	GetTotalViewers(defaultViewers int) (int, *errs.Error)
	SetTotalViewers(totalViewers int) *errs.Error
}

type MaxViewersActivator struct {
	store        MaxViewersActivatorStore
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

	if m.totalViewers+1 > m.maxViewers {
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
	store MaxViewersActivatorStore,
	maxViewers int,
) (*MaxViewersActivator, *errs.Error) {
	totalViewers, err := store.GetTotalViewers(0)
	if err != nil {
		return nil, err
	}

	return &MaxViewersActivator{
		store:        store,
		totalViewers: totalViewers,
		maxViewers:   maxViewers,
	}, nil
}

type PercentageActivatorStore interface {
	GetIsActivated(viewerID uint64) (*bool, *errs.Error)
	SetIsActivated(viewerID uint64, isActivated bool) *errs.Error
}

type PercentageActivator struct {
	store      PercentageActivatorStore
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

	randInt := p.randomGen.RandomInt(100)
	isActive := randInt < p.percentage
	err = p.store.SetIsActivated(viewerID, isActive)
	return isActive, err
}

func NewPercentageActivator(
	store PercentageActivatorStore,
	randomGen randgen.RandomNumberGenerator,
	percentage int,
) *PercentageActivator {
	return &PercentageActivator{
		store:      store,
		randomGen:  randomGen,
		percentage: percentage,
	}
}

type Bucket struct {
	Percentage      int
	MinimalBakeTime time.Duration
}

type IncrementalPercentageActivatorStore interface {
	GetIsActivated(viewerID uint64) (*bool, *errs.Error)
	SetIsActivated(viewerID uint64, isActivated bool) *errs.Error
	GetBucketIndex(defaultBucketIndex int) (int, *errs.Error)
	SetBucketIndex(bucketIndex int) *errs.Error
}

type IncrementalPercentageActivator struct {
	store         IncrementalPercentageActivatorStore
	randomGen     randgen.RandomNumberGenerator
	clock         clock.Clock
	buckets       []Bucket
	bucketIndex   int
	bucketStartAt time.Time
}

var _ Activator = (*IncrementalPercentageActivator)(nil)

func (i *IncrementalPercentageActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	//TODO: fix bug that some viewers are not activated forever
	isActivated, err := i.store.GetIsActivated(viewerID)
	if err != nil {
		return false, err
	}

	if isActivated != nil {
		return *isActivated, nil
	}

	now := i.clock.Now().UTC()
	for i.bucketIndex < len(i.buckets) {
		timeElapsed := now.Sub(i.bucketStartAt)
		if timeElapsed < i.buckets[i.bucketIndex].MinimalBakeTime {
			break
		}

		i.bucketStartAt = i.bucketStartAt.Add(i.buckets[i.bucketIndex].MinimalBakeTime)
		i.bucketIndex++
		err = i.store.SetBucketIndex(i.bucketIndex)
		if err != nil {
			return false, err
		}
	}

	if i.bucketIndex >= len(i.buckets) {
		return true, nil
	}

	randInt := i.randomGen.RandomInt(100)
	isActive := randInt < i.buckets[i.bucketIndex].Percentage
	err = i.store.SetIsActivated(viewerID, isActive)
	return isActive, err
}

func NewIncrementalPercentageActivator(
	store IncrementalPercentageActivatorStore,
	randomGen randgen.RandomNumberGenerator,
	clock clock.Clock,
	buckets []Bucket,
) (*IncrementalPercentageActivator, *errs.Error) {
	bucketIndex, err := store.GetBucketIndex(0)
	if err != nil {
		return nil, err
	}

	return &IncrementalPercentageActivator{
		clock:         clock,
		store:         store,
		randomGen:     randomGen,
		buckets:       buckets,
		bucketIndex:   bucketIndex,
		bucketStartAt: clock.Now().UTC(),
	}, nil
}

type ChainedActivator struct {
	activators []Activator
}

var _ Activator = (*ChainedActivator)(nil)

func (c *ChainedActivator) IsActive(viewerID uint64) (bool, *errs.Error) {
	for _, activator := range c.activators {
		isActive, err := activator.IsActive(viewerID)
		if err != nil {
			return false, err
		}

		if !isActive {
			return false, nil
		}
	}

	return true, nil
}

func NewChainedActivator(activators []Activator) *ChainedActivator {
	return &ChainedActivator{
		activators: activators,
	}
}
