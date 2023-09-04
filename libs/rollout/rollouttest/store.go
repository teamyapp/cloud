package rollouttest

import (
	"context"
	"maps"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/rollout"
)

type Viewer struct {
	VersionNumber int
	IsActivated   bool
}

type Store struct {
	viewers          map[uint64]*Viewer
	totalViewers     *int
	isRolloutEnabled *bool
	bucketIndex      int
}

var _ rollout.Store = (*Store)(nil)

func (s *Store) GetViewerVersionNumber(ct context.Context, viewerID uint64) (*int, *errs.Error) {
	viewer, ok := s.viewers[viewerID]
	if !ok {
		return nil, nil
	}

	return &viewer.VersionNumber, nil
}

func (s *Store) SetViewerVersionNumber(ct context.Context, viewerID uint64, versionNumber int) *errs.Error {
	viewer, ok := s.viewers[viewerID]
	if !ok {
		viewer = &Viewer{}
		s.viewers[viewerID] = viewer
	}

	viewer.VersionNumber = versionNumber
	return nil
}

func (s *Store) GetTotalViewers(ct context.Context, defaultViewers int) (int, *errs.Error) {
	if s.totalViewers == nil {
		return defaultViewers, nil
	}

	return *s.totalViewers, nil
}

func (s *Store) SetTotalViewers(ct context.Context, totalViewers int) *errs.Error {
	s.totalViewers = &totalViewers
	return nil
}

func (s *Store) GetIsActivated(ct context.Context, viewerID uint64) (*bool, *errs.Error) {
	viewer, ok := s.viewers[viewerID]
	if !ok {
		return nil, nil
	}

	return &viewer.IsActivated, nil
}

func (s *Store) SetIsActivated(ct context.Context, viewerID uint64, isActivated bool) *errs.Error {
	viewer, ok := s.viewers[viewerID]
	if !ok {
		viewer = &Viewer{}
		s.viewers[viewerID] = viewer
	}

	viewer.IsActivated = isActivated
	return nil
}

func (s *Store) GetIsRolloutEnabled(ct context.Context, defaultIsRolloutEnabled bool) (bool, *errs.Error) {
	if s.isRolloutEnabled == nil {
		return defaultIsRolloutEnabled, nil
	}

	return *s.isRolloutEnabled, nil
}

func (s *Store) SetIsRolloutEnabled(ct context.Context, isRolloutEnabled bool) *errs.Error {
	s.isRolloutEnabled = &isRolloutEnabled
	return nil
}

func (s *Store) GetBucketIndex(ct context.Context, defaultBucketIndex int) (int, *errs.Error) {
	return s.bucketIndex, nil
}

func (s *Store) SetBucketIndex(ct context.Context, bucketIndex int) *errs.Error {
	s.bucketIndex = bucketIndex
	return nil
}

type StoreBuilder struct {
	viewers          map[uint64]*Viewer
	totalViewers     *int
	isRolloutEnabled *bool
	bucketIndex      int
}

func (b *StoreBuilder) WithViewers(viewers map[uint64]*Viewer) *StoreBuilder {
	b.viewers = maps.Clone(viewers)
	return b
}

func (b *StoreBuilder) WithTotalViewers(totalViewers int) *StoreBuilder {
	b.totalViewers = &totalViewers
	return b
}

func (b *StoreBuilder) WithIsRolloutEnabled(isRolloutEnabled bool) *StoreBuilder {
	b.isRolloutEnabled = &isRolloutEnabled
	return b
}

func (b *StoreBuilder) WithBucketIndex(bucketIndex int) *StoreBuilder {
	b.bucketIndex = bucketIndex
	return b
}

func (b *StoreBuilder) Build() *Store {
	return &Store{
		viewers:          b.viewers,
		totalViewers:     b.totalViewers,
		isRolloutEnabled: b.isRolloutEnabled,
		bucketIndex:      b.bucketIndex,
	}
}

func NewStoreBuilder() *StoreBuilder {
	return &StoreBuilder{
		viewers: map[uint64]*Viewer{},
	}
}
