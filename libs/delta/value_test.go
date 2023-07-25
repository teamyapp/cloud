package delta

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectValueDelta(t *testing.T) {
	oldValue := "old"
	newValue := "new"
	expectedDelta := Delta[string]{
		Status: UpdatedStatus,
		Value:  newValue,
	}

	delta := DetectValueDelta(oldValue, newValue)
	require.Equal(t, expectedDelta, delta)
}

func TestToValueDelta(t *testing.T) {
	value := "value"
	valueDelta := ToValueDelta(UpdatedStatus, value)
	require.Equal(t, value, valueDelta)
}
