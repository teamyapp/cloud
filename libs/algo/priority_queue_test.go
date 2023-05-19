package algo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/libs/errs"
)

func TestPriorityQuery_Int(t *testing.T) {
	testCases := []struct {
		name     string
		compare  Comparator[int]
		initials []int
		inserts  []int
		peaks    []int
		tops     []int
	}{
		{
			name:     "min priority queue",
			compare:  CompareAsc[int],
			initials: []int{},
			inserts:  []int{1, 3, 5, 2, 4, 0},
			peaks:    []int{1, 1, 1, 1, 1, 0},
			tops:     []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "max priority queue",
			compare:  CompareDesc[int],
			initials: []int{},
			inserts:  []int{1, 3, 5, 2, 4, 0},
			peaks:    []int{1, 3, 5, 5, 5, 5},
			tops:     []int{5, 4, 3, 2, 1, 0},
		},
		{
			name:     "min priority queue with initial values",
			compare:  CompareAsc[int],
			initials: []int{6, 3, 5, 2, 4, 1},
			inserts:  []int{7, 0},
			peaks:    []int{1, 0},
			tops:     []int{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:     "min priority queue shift down to pop and find higher priority node in the left child",
			compare:  CompareAsc[int],
			initials: []int{6, 3, 5, 2, 4, 1},
			inserts:  []int{},
			peaks:    []int{},
			tops:     []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "min priority queue shift down to pop and find higher priority node in the right child",
			compare:  CompareAsc[int],
			initials: []int{6, 3, 5, 4, 2, 1},
			inserts:  []int{},
			peaks:    []int{},
			tops:     []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "min priority queue push element needs to shift up",
			compare:  CompareAsc[int],
			initials: []int{6, 3, 5, 4, 2, 1},
			inserts:  []int{1},
			peaks:    []int{1},
			tops:     []int{1, 1, 2, 3, 4, 5, 6},
		},
		{
			name:     "min priority queue push element without shift up",
			compare:  CompareAsc[int],
			initials: []int{6, 3, 5, 4, 2, 1},
			inserts:  []int{7},
			peaks:    []int{1},
			tops:     []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:     "min priority queue with negative values",
			compare:  CompareAsc[int],
			initials: []int{6, 3, -5, 4, 2, 1},
			inserts:  []int{-7, -8},
			peaks:    []int{-7, -8},
			tops:     []int{-8, -7, -5, 1, 2, 3, 4, 6},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pq := NewPriorityQueue(
				testCase.compare,
				testCase.initials,
			)

			assert.Equal(t, len(testCase.initials), pq.Size())
			initialSize := pq.Size()

			for index, insert := range testCase.inserts {
				pq.Push(insert)
				assert.Equal(t, index+1+initialSize, pq.Size())
				value, _ := pq.Peek()
				assert.Equal(t, testCase.peaks[index], value)
			}

			for _, top := range testCase.tops {
				cur_size := pq.Size()
				value, _ := pq.Peek()
				assert.Equal(t, top, value)
				value, _ = pq.Pop()
				assert.Equal(t, top, value)
				assert.Equal(t, cur_size-1, pq.Size())
			}

			assert.Zero(t, pq.Size())

			value, err := pq.Peek()
			assert.Equal(t, err.Code, errs.InvalidOperation)
			assert.Equal(t, 0, value)
			value, err = pq.Pop()
			assert.Equal(t, 0, value)
			assert.Equal(t, err.Code, errs.InvalidOperation)
		})
	}

}

func TestPriorityQuery_String(t *testing.T) {
	testCases := []struct {
		name     string
		compare  Comparator[string]
		initials []string
		inserts  []string
		peaks    []string
		tops     []string
	}{
		{
			name:     "min priority queue",
			compare:  CompareAsc[string],
			initials: []string{},
			inserts: []string{
				"1",
				"3",
				"5",
				"2",
				"4",
				"0",
			},
			peaks: []string{"1", "1", "1", "1", "1", "0"},
			tops:  []string{"0", "1", "2", "3", "4", "5"},
		},
		{
			name:     "max priority queue",
			compare:  CompareDesc[string],
			initials: []string{},
			inserts: []string{
				"1",
				"3",
				"5",
				"2",
				"4",
				"0",
			},
			peaks: []string{"1", "3", "5", "5", "5", "5"},
			tops:  []string{"5", "4", "3", "2", "1", "0"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pq := NewPriorityQueue(
				testCase.compare,
				testCase.initials,
			)

			assert.Equal(t, len(testCase.initials), pq.Size())

			for index, insert := range testCase.inserts {
				pq.Push(insert)
				assert.Equal(t, index+1, pq.Size())
				value, _ := pq.Peek()
				assert.Equal(t, testCase.peaks[index], value)
			}

			for index, top := range testCase.tops {
				value, _ := pq.Peek()
				assert.Equal(t, top, value)
				value, _ = pq.Pop()
				assert.Equal(t, top, value)
				assert.Equal(t, len(testCase.tops)-index-1, pq.Size())
			}

			assert.Zero(t, pq.Size())

			value, err := pq.Peek()
			assert.Equal(t, err.Code, errs.InvalidOperation)
			assert.Equal(t, "", value)
			value, err = pq.Pop()
			assert.Equal(t, "", value)
			assert.Equal(t, err.Code, errs.InvalidOperation)
		})
	}

}

func TestPriorityQuery_Date(t *testing.T) {
	testCases := []struct {
		name     string
		compare  func(value1, value2 time.Time) Comparison
		initials []time.Time
		inserts  []time.Time
		peaks    []time.Time
		tops     []time.Time
	}{
		{
			name: "min priority queue",
			compare: func(time1 time.Time, time2 time.Time) Comparison {
				if time1.After(time2) {
					return GreaterThan
				}

				if time1.Before(time2) {
					return SmallerThan
				}

				return Equal
			},
			initials: []time.Time{},
			inserts: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			peaks: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			tops: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
			},
		},
		{
			name: "max priority queue",
			compare: func(time1 time.Time, time2 time.Time) Comparison {
				if time1.After(time2) {
					return SmallerThan
				}

				if time1.Before(time2) {
					return GreaterThan
				}

				return Equal
			},
			initials: []time.Time{},
			inserts: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			peaks: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
			},
			tops: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pq := NewPriorityQueue(
				testCase.compare,
				testCase.initials,
			)

			assert.Equal(t, len(testCase.initials), pq.Size())

			for index, insert := range testCase.inserts {
				pq.Push(insert)
				assert.Equal(t, index+1, pq.Size())
				value, _ := pq.Peek()
				assert.Equal(t, testCase.peaks[index], value)
			}

			for index, top := range testCase.tops {
				value, _ := pq.Peek()
				assert.Equal(t, top, value)
				value, _ = pq.Pop()
				assert.Equal(t, top, value)
				assert.Equal(t, len(testCase.tops)-index-1, pq.Size())
			}

			assert.Zero(t, pq.Size())

			value, err := pq.Peek()
			assert.Equal(t, err.Code, errs.InvalidOperation)
			assert.Equal(t, time.Time(time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)), value)
			value, err = pq.Pop()
			assert.Equal(t, time.Time(time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)), value)
			assert.Equal(t, err.Code, errs.InvalidOperation)
		})
	}
}
