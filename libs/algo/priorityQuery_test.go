package algo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQuery_Int(t *testing.T) {

	testCases := []struct {
		name       string
		comparator func(priority1, priority2 int) bool
		inserts    []int
		peaks      []int
		tops       []int
	}{
		{
			name: "min priority queue",
			comparator: func(priority1, priority2 int) bool {
				return priority1 < priority2
			},
			inserts: []int{1, 3, 5, 2, 4, 0},
			peaks:   []int{1, 1, 1, 1, 1, 0},
			tops:    []int{0, 1, 2, 3, 4, 5},
		},
		{
			name: "max priority queue",
			comparator: func(priority1, priority2 int) bool {
				return priority1 > priority2
			},
			inserts: []int{1, 3, 5, 2, 4, 0},
			peaks:   []int{1, 3, 5, 5, 5, 5},
			tops:    []int{5, 4, 3, 2, 1, 0},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pq := NewPriorityQueue[int](
				testCase.comparator,
			)

			assert.Equal(t, 0, pq.Size())

			for index, insert := range testCase.inserts {
				pq.Push(insert, insert)
				assert.Equal(t, index+1, pq.Size())
				assert.Equal(t, testCase.peaks[index], *pq.Peek())
			}

			for index, top := range testCase.tops {
				assert.Equal(t, top, *pq.Peek())
				assert.Equal(t, top, *pq.Pop())
				assert.Equal(t, len(testCase.tops)-index-1, pq.Size())
			}

			assert.Zero(t, pq.Size())
			assert.Nil(t, pq.Peek())
			assert.Nil(t, pq.Pop())
		})
	}

}

func TestPriorityQuery_String(t *testing.T) {

	testCases := []struct {
		name       string
		comparator func(priority1, priority2 int) bool
		inserts    []struct {
			value    string
			priority int
		}
		peaks []string
		tops  []string
	}{
		{
			name: "min priority queue",
			comparator: func(priority1, priority2 int) bool {
				return priority1 < priority2
			},
			inserts: []struct {
				value    string
				priority int
			}{
				{"1", 1},
				{"3", 3},
				{"5", 5},
				{"2", 2},
				{"4", 4},
				{"0", 0},
			},
			peaks: []string{"1", "1", "1", "1", "1", "0"},
			tops:  []string{"0", "1", "2", "3", "4", "5"},
		},
		{
			name: "max priority queue",
			comparator: func(priority1, priority2 int) bool {
				return priority1 > priority2
			},
			inserts: []struct {
				value    string
				priority int
			}{
				{"1", 1},
				{"3", 3},
				{"5", 5},
				{"2", 2},
				{"4", 4},
				{"0", 0},
			},
			peaks: []string{"1", "3", "5", "5", "5", "5"},
			tops:  []string{"5", "4", "3", "2", "1", "0"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pq := NewPriorityQueue[string](
				testCase.comparator,
			)

			assert.Equal(t, 0, pq.Size())

			for index, insert := range testCase.inserts {
				pq.Push(insert.value, insert.priority)
				assert.Equal(t, index+1, pq.Size())
				assert.Equal(t, testCase.peaks[index], *pq.Peek())
			}

			for index, top := range testCase.tops {
				assert.Equal(t, top, *pq.Peek())
				assert.Equal(t, top, *pq.Pop())
				assert.Equal(t, len(testCase.tops)-index-1, pq.Size())
			}

			assert.Zero(t, pq.Size())
			assert.Nil(t, pq.Peek())
			assert.Nil(t, pq.Pop())
		})

	}

}
