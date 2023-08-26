package rollout_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/randgen/randgentest"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/cloud/libs/rollout/rollouttest"
)

type Deps struct {
	store   rollout.Store
	clock   clock.Clock
	randGen randgen.RandomNumberGenerator
}

func TestOrderedRollouts_GetVersionNumber(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name                   string
		makeRollouts           func(deps Deps) rollout.OrderedRollouts
		nows                   []time.Time
		randInts               []int
		viewerIDs              []uint64
		expectedVersionNumbers []*int
	}{
		{
			name: "no rollouts",
			makeRollouts: func(deps Deps) rollout.OrderedRollouts {
				return rollout.OrderedRollouts{}
			},
			nows:                   []time.Time{now, now, now},
			viewerIDs:              []uint64{1, 2, 3},
			expectedVersionNumbers: []*int{nil, nil, nil},
		},
		{
			name: "static rollouts",
			makeRollouts: func(deps Deps) rollout.OrderedRollouts {
				rollout1, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(1),
				)
				require.Nil(t, err)

				rollout2, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(2),
				)
				require.Nil(t, err)

				rollout3, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(3),
				)

				return []rollout.Rollout{rollout1, rollout2, rollout3}
			},
			nows:                   []time.Time{now, now, now},
			viewerIDs:              []uint64{1, 2, 3},
			expectedVersionNumbers: []*int{intPtr(3), intPtr(3), intPtr(3)},
		},
		{
			name: "time range rollouts",
			makeRollouts: func(deps Deps) rollout.OrderedRollouts {
				rollout1, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(1),
				)
				require.Nil(t, err)

				startAt1 := now.Add(time.Hour)
				endAt1 := now.Add(4 * time.Hour)
				rollout2, err := rollout.NewRollout(
					deps.store,
					rollout.NewTimeRangeActivator(deps.clock, &startAt1, &endAt1),
					rollout.NewStaticVersionSelector(2),
				)
				require.Nil(t, err)

				startAt2 := now.Add(2 * time.Hour)
				endAt2 := now.Add(3 * time.Hour)
				rollout3, err := rollout.NewRollout(
					deps.store,
					rollout.NewTimeRangeActivator(deps.clock, &startAt2, &endAt2),
					rollout.NewStaticVersionSelector(3),
				)
				require.Nil(t, err)

				return []rollout.Rollout{rollout1, rollout2, rollout3}
			},
			nows: []time.Time{
				now.Add(30 * time.Minute),
				now.Add(1*time.Hour + 30*time.Minute),
				now.Add(2*time.Hour + 30*time.Minute),
				now.Add(3*time.Hour + 30*time.Minute),
				now.Add(4*time.Hour + 30*time.Minute),
			},
			viewerIDs:              []uint64{1, 2, 3, 2, 1},
			expectedVersionNumbers: []*int{intPtr(1), intPtr(2), intPtr(3), intPtr(2), intPtr(1)},
		},
		{
			name: "max viewers rollouts",
			makeRollouts: func(deps Deps) rollout.OrderedRollouts {
				rollout1, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(1),
				)
				require.Nil(t, err)

				maxViewerActivator, err := rollout.NewMaxViewersActivator(deps.store, 2)
				require.Nil(t, err)

				rollout2, err := rollout.NewRollout(
					deps.store,
					maxViewerActivator,
					rollout.NewStaticVersionSelector(2),
				)
				require.Nil(t, err)

				return []rollout.Rollout{rollout1, rollout2}
			},
			nows:                   []time.Time{now, now, now, now, now},
			viewerIDs:              []uint64{1, 2, 3, 1, 2},
			expectedVersionNumbers: []*int{intPtr(2), intPtr(2), intPtr(1), intPtr(2), intPtr(2)},
		},
		{
			name: "percentage rollouts",
			makeRollouts: func(deps Deps) rollout.OrderedRollouts {
				rollout1, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(1),
				)
				require.Nil(t, err)

				percentageActivator := rollout.NewPercentageActivator(deps.store, deps.randGen, 40)
				require.Nil(t, err)

				rollout2, err := rollout.NewRollout(
					deps.store,
					percentageActivator,
					rollout.NewStaticVersionSelector(2),
				)
				require.Nil(t, err)

				return []rollout.Rollout{rollout1, rollout2}
			},
			nows:      []time.Time{now, now, now, now, now, now},
			randInts:  []int{10, 20, 50},
			viewerIDs: []uint64{1, 2, 5, 5, 2, 1},
			expectedVersionNumbers: []*int{
				intPtr(2),
				intPtr(2),
				intPtr(1),
				intPtr(1),
				intPtr(2),
				intPtr(2),
			},
		},
		{
			name: "incremental percentage rollouts",
			makeRollouts: func(deps Deps) rollout.OrderedRollouts {
				rollout1, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(1),
				)
				require.Nil(t, err)

				incrementalPercentageActivator, err := rollout.NewIncrementalPercentageActivator(
					deps.store,
					deps.randGen,
					deps.clock,
					[]rollout.Bucket{
						{
							Percentage:      10,
							MinimalBakeTime: 1 * time.Hour,
						},
						{
							Percentage:      30,
							MinimalBakeTime: 3 * time.Hour,
						},
						{
							Percentage:      70,
							MinimalBakeTime: 5 * time.Hour,
						},
					},
				)
				require.Nil(t, err)
				rollout2, err := rollout.NewRollout(
					deps.store,
					incrementalPercentageActivator,
					rollout.NewStaticVersionSelector(2),
				)
				require.Nil(t, err)
				return []rollout.Rollout{rollout1, rollout2}
			},
			nows: []time.Time{
				now,                                   /* bucket 1 */
				now.Add(30 * time.Minute),             /* bucket 1 */
				now.Add(40 * time.Minute),             /* bucket 1 */
				now.Add(1 * time.Hour),                /* bucket 2 */
				now.Add(2 * time.Hour),                /* bucket 2 */
				now.Add(2*time.Hour + 30*time.Minute), /* bucket 2 */
				now.Add(12 * time.Hour),               /* bucket 3 */
			},
			randInts: []int{
				5,  /* bucket 1 */
				25, /* bucket 1 */
				25, /* bucket 2 */
				50, /* bucket 2 */
				80, /* bucket 3 */
			},
			viewerIDs: []uint64{
				1, /* bucket 1 */
				2, /* bucket 1 */
				1, /* bucket 1 */
				3, /* bucket 2 */
				2, /* bucket 2 */
				4, /* bucket 2 */
				5, /* bucket 3 */
			},
			expectedVersionNumbers: []*int{
				intPtr(2), /* viewer 1 */
				intPtr(1), /* viewer 2 */
				intPtr(2), /* viewer 1 */
				intPtr(2), /* viewer 3 */
				intPtr(1), /* viewer 2 */
				intPtr(1), /* viewer 4 */
				intPtr(2), /* viewer 5 */
			},
		},
		{
			name: "experiment rollouts",
			makeRollouts: func(deps Deps) rollout.OrderedRollouts {
				rollout1, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewStaticVersionSelector(1),
				)
				require.Nil(t, err)

				rollout2, err := rollout.NewRollout(
					deps.store,
					rollout.NewStaticActivator(),
					rollout.NewExperimentVersionSelector(
						deps.store,
						deps.randGen,
						[]int{2, 3},
					),
				)
				require.Nil(t, err)

				return []rollout.Rollout{rollout1, rollout2}
			},
			nows:      []time.Time{now, now, now, now, now},
			randInts:  []int{0, 0, 1},
			viewerIDs: []uint64{1, 2, 3, 2, 3},
			expectedVersionNumbers: []*int{
				intPtr(2),
				intPtr(2),
				intPtr(3),
				intPtr(2),
				intPtr(3),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := rollouttest.NewStoreBuilder().Build()
			mockClock := clock.NewMock()
			mockClock.Set(now)
			stubRandGen := randgentest.NewStubRanGen(testCase.randInts)
			deps := Deps{
				store:   store,
				clock:   mockClock,
				randGen: stubRandGen,
			}
			rollouts := testCase.makeRollouts(deps)
			for index, viewerID := range testCase.viewerIDs {
				mockClock.Set(testCase.nows[index])
				versionNumber, err := rollouts.GetVersionNumber(viewerID)
				require.Nil(t, err, "index: %d", index)

				if testCase.expectedVersionNumbers[index] == nil {
					require.Nil(t, versionNumber, "index: %d", index)
				} else {
					require.Equal(
						t,
						*testCase.expectedVersionNumbers[index],
						*versionNumber,
						"index: %d", index)
				}
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
