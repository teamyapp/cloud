package delta_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/libs/delta"
	"github.com/teamyapp/cloud/libs/delta/deltatest"
)

func TestDetectMapDeltaSimpleValue(t *testing.T) {
	testCase := []struct {
		name          string
		old           map[string]string
		new           map[string]string
		expectedDelta delta.Delta[map[string]delta.KeyValueDelta[string]]
	}{
		{
			name: "map unchanged",
			old: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			new: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[string]]{
				Status: delta.UnchangedStatus,
				Value: map[string]delta.KeyValueDelta[string]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value:       "value1",
					},
					"key2": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value:       "value2",
					},
				},
			},
		},
		{
			name: "added key",
			old: map[string]string{
				"key1": "value1",
			},
			new: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[string]]{
				Status: delta.UpdatedStatus,
				Value: map[string]delta.KeyValueDelta[string]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value:       "value1",
					},
					"key2": {
						KeyStatus:   delta.AddedStatus,
						ValueStatus: delta.AddedStatus,
						Value:       "value2",
					},
				},
			},
		},
		{
			name: "updated value",
			old: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			new: map[string]string{
				"key1": "value1",
				"key2": "value3",
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[string]]{
				Status: delta.UpdatedStatus,
				Value: map[string]delta.KeyValueDelta[string]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value:       "value1",
					},
					"key2": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UpdatedStatus,
						Value:       "value3",
					},
				},
			},
		},
		{
			name: "removed key",
			old: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			new: map[string]string{
				"key1": "value1",
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[string]]{
				Status: delta.UpdatedStatus,
				Value: map[string]delta.KeyValueDelta[string]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value:       "value1",
					},
					"key2": {
						KeyStatus:   delta.RemovedStatus,
						ValueStatus: delta.RemovedStatus,
						Value:       "value2",
					},
				},
			},
		},
	}

	for _, tc := range testCase {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dt := delta.DetectMapDelta(tc.old, tc.new, delta.DetectValueDelta[string], delta.ToValueDelta[string])
			require.Equal(t, tc.expectedDelta, dt)
		})
	}
}

func TestDetectMapDeltaNestedDelta(t *testing.T) {
	testCase := []struct {
		name          string
		old           map[string]deltatest.NestedValue
		new           map[string]deltatest.NestedValue
		expectedDelta delta.Delta[map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]]
	}{
		{
			name: "map unchanged",
			old: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   false,
				},
				"key2": {
					FirstName: "Jane",
					LastName:  "Ashley",
					Age:       28,
					IsAdmin:   true,
				},
			},
			new: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   false,
				},
				"key2": {

					FirstName: "Jane",
					LastName:  "Ashley",
					Age:       28,
					IsAdmin:   true,
				},
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]]{
				Status: delta.UnchangedStatus,
				Value: map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.UnchangedStatus, Value: "John"},
							LastName:  delta.Delta[string]{Status: delta.UnchangedStatus, Value: "Doe"},
							Age:       delta.Delta[int]{Status: delta.UnchangedStatus, Value: 30},
							IsAdmin:   delta.Delta[bool]{Status: delta.UnchangedStatus, Value: false},
						},
					},
					"key2": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.UnchangedStatus, Value: "Jane"},
							LastName:  delta.Delta[string]{Status: delta.UnchangedStatus, Value: "Ashley"},
							Age:       delta.Delta[int]{Status: delta.UnchangedStatus, Value: 28},
							IsAdmin:   delta.Delta[bool]{Status: delta.UnchangedStatus, Value: true},
						},
					},
				},
			},
		},
		{
			name: "added key",
			old: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   false,
				},
			},
			new: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   false,
				},
				"key2": {
					FirstName: "Jane",
					LastName:  "Ashley",
					Age:       28,
					IsAdmin:   true,
				},
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]]{
				Status: delta.UpdatedStatus,
				Value: map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.UnchangedStatus, Value: "John"},
							LastName:  delta.Delta[string]{Status: delta.UnchangedStatus, Value: "Doe"},
							Age:       delta.Delta[int]{Status: delta.UnchangedStatus, Value: 30},
							IsAdmin:   delta.Delta[bool]{Status: delta.UnchangedStatus, Value: false},
						},
					},
					"key2": {
						KeyStatus:   delta.AddedStatus,
						ValueStatus: delta.AddedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.AddedStatus, Value: "Jane"},
							LastName:  delta.Delta[string]{Status: delta.AddedStatus, Value: "Ashley"},
							Age:       delta.Delta[int]{Status: delta.AddedStatus, Value: 28},
							IsAdmin:   delta.Delta[bool]{Status: delta.AddedStatus, Value: true},
						},
					},
				},
			},
		},
		{
			name: "updated nest value",
			old: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   false,
				},
				"key2": {
					FirstName: "Jane",
					LastName:  "Ashley",
					Age:       28,
					IsAdmin:   true,
				},
			},
			new: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   true,
				},
				"key2": {
					FirstName: "Jane",
					LastName:  "Ellis",
					Age:       29,
					IsAdmin:   true,
				},
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]]{
				Status: delta.UpdatedStatus,
				Value: map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UpdatedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.UnchangedStatus, Value: "John"},
							LastName:  delta.Delta[string]{Status: delta.UnchangedStatus, Value: "Doe"},
							Age:       delta.Delta[int]{Status: delta.UnchangedStatus, Value: 30},
							IsAdmin:   delta.Delta[bool]{Status: delta.UpdatedStatus, Value: true},
						},
					},
					"key2": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UpdatedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.UnchangedStatus, Value: "Jane"},
							LastName:  delta.Delta[string]{Status: delta.UpdatedStatus, Value: "Ellis"},
							Age:       delta.Delta[int]{Status: delta.UpdatedStatus, Value: 29},
							IsAdmin:   delta.Delta[bool]{Status: delta.UnchangedStatus, Value: true},
						},
					},
				},
			},
		},
		{
			name: "removed key",
			old: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   false,
				},
				"key2": {
					FirstName: "Jane",
					LastName:  "Ashley",
					Age:       28,
					IsAdmin:   true,
				},
			},
			new: map[string]deltatest.NestedValue{
				"key1": {
					FirstName: "John",
					LastName:  "Doe",
					Age:       30,
					IsAdmin:   false,
				},
			},
			expectedDelta: delta.Delta[map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]]{
				Status: delta.UpdatedStatus,
				Value: map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]{
					"key1": {
						KeyStatus:   delta.UnchangedStatus,
						ValueStatus: delta.UnchangedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.UnchangedStatus, Value: "John"},
							LastName:  delta.Delta[string]{Status: delta.UnchangedStatus, Value: "Doe"},
							Age:       delta.Delta[int]{Status: delta.UnchangedStatus, Value: 30},
							IsAdmin:   delta.Delta[bool]{Status: delta.UnchangedStatus, Value: false},
						},
					},
					"key2": {
						KeyStatus:   delta.RemovedStatus,
						ValueStatus: delta.RemovedStatus,
						Value: deltatest.NestedValueDelta{
							FirstName: delta.Delta[string]{Status: delta.RemovedStatus, Value: "Jane"},
							LastName:  delta.Delta[string]{Status: delta.RemovedStatus, Value: "Ashley"},
							Age:       delta.Delta[int]{Status: delta.RemovedStatus, Value: 28},
							IsAdmin:   delta.Delta[bool]{Status: delta.RemovedStatus, Value: true},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCase {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dt := delta.DetectMapDelta(tc.old, tc.new, deltatest.DetectNestedValueDelta, deltatest.ToNestValueDelta)
			require.Equal(t, tc.expectedDelta, dt)
		})
	}
}

func TestToMapDeltaSimpleValue(t *testing.T) {
	inputMap := map[string]int{
		"key1": 1,
		"key2": 2,
	}
	status := delta.AddedStatus
	expectedDelta := map[string]delta.KeyValueDelta[int]{
		"key1": {
			KeyStatus:   delta.AddedStatus,
			ValueStatus: delta.AddedStatus,
			Value:       1,
		},
		"key2": {
			KeyStatus:   delta.AddedStatus,
			ValueStatus: delta.AddedStatus,
			Value:       2,
		},
	}
	dt := delta.ToMapDelta(status, inputMap, delta.ToValueDelta[int])
	require.Equal(t, expectedDelta, dt)
}

func TestToMapDeltaNestDelta(t *testing.T) {
	inputMap := map[string]deltatest.NestedValue{
		"key1": {
			FirstName: "John",
			LastName:  "Doe",
			Age:       30,
			IsAdmin:   false,
		},
		"key2": {
			FirstName: "Jane",
			LastName:  "Ashley",
			Age:       28,
			IsAdmin:   true,
		},
	}
	status := delta.AddedStatus
	expectedDelta := map[string]delta.KeyValueDelta[deltatest.NestedValueDelta]{
		"key1": {
			KeyStatus:   delta.AddedStatus,
			ValueStatus: delta.AddedStatus,
			Value: deltatest.NestedValueDelta{
				FirstName: delta.Delta[string]{Status: delta.AddedStatus, Value: "John"},
				LastName:  delta.Delta[string]{Status: delta.AddedStatus, Value: "Doe"},
				Age:       delta.Delta[int]{Status: delta.AddedStatus, Value: 30},
				IsAdmin:   delta.Delta[bool]{Status: delta.AddedStatus, Value: false},
			},
		},
		"key2": {
			KeyStatus:   delta.AddedStatus,
			ValueStatus: delta.AddedStatus,
			Value: deltatest.NestedValueDelta{
				FirstName: delta.Delta[string]{Status: delta.AddedStatus, Value: "Jane"},
				LastName:  delta.Delta[string]{Status: delta.AddedStatus, Value: "Ashley"},
				Age:       delta.Delta[int]{Status: delta.AddedStatus, Value: 28},
				IsAdmin:   delta.Delta[bool]{Status: delta.AddedStatus, Value: true},
			},
		},
	}
	dt := delta.ToMapDelta(status, inputMap, deltatest.ToNestValueDelta)
	require.Equal(t, expectedDelta, dt)
}
