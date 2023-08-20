package algo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/libs/errs"
)

func TestPriorityQuery_Int(t *testing.T) {
	testCases := []struct {
		name          string
		compare       Comparator[int]
		initialValues []int
		inserts       []int
		peeks         []int
		pops          []int
	}{
		{
			name:          "min priority queue",
			compare:       CompareAsc[int],
			initialValues: []int{},
			inserts:       []int{1, 3, 5, 2, 4, 0},
			peeks:         []int{1, 1, 1, 1, 1, 0},
			pops:          []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:          "max priority queue",
			compare:       CompareDesc[int],
			initialValues: []int{},
			inserts:       []int{1, 3, 5, 2, 4, 0},
			peeks:         []int{1, 3, 5, 5, 5, 5},
			pops:          []int{5, 4, 3, 2, 1, 0},
		},
		{
			name:          "min priority queue with initial values",
			compare:       CompareAsc[int],
			initialValues: []int{6, 3, 5, 2, 4, 1},
			inserts:       []int{7, 0},
			peeks:         []int{1, 0},
			pops:          []int{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:          "min priority queue shift down to pop and find higher priority node in the left child",
			compare:       CompareAsc[int],
			initialValues: []int{6, 3, 5, 2, 4, 1},
			inserts:       []int{},
			peeks:         []int{},
			pops:          []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:          "min priority queue shift down to pop and find higher priority node in the right child",
			compare:       CompareAsc[int],
			initialValues: []int{6, 3, 5, 4, 2, 1},
			inserts:       []int{},
			peeks:         []int{},
			pops:          []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:          "min priority queue push element needs to shift up",
			compare:       CompareAsc[int],
			initialValues: []int{6, 3, 5, 4, 2, 1},
			inserts:       []int{1},
			peeks:         []int{1},
			pops:          []int{1, 1, 2, 3, 4, 5, 6},
		},
		{
			name:          "min priority queue push element without shift up",
			compare:       CompareAsc[int],
			initialValues: []int{6, 3, 5, 4, 2, 1},
			inserts:       []int{7},
			peeks:         []int{1},
			pops:          []int{1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:          "min priority queue with negative values",
			compare:       CompareAsc[int],
			initialValues: []int{6, 3, -5, 4, 2, 1},
			inserts:       []int{-7, -8},
			peeks:         []int{-7, -8},
			pops:          []int{-8, -7, -5, 1, 2, 3, 4, 6},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pq := NewPriorityQueue(
				testCase.compare,
				testCase.initialValues,
			)

			require.Equal(t, len(testCase.initialValues), pq.Size())
			initialSize := pq.Size()

			for index, insert := range testCase.inserts {
				pq.Insert(insert)
				require.Equal(t, index+1+initialSize, pq.Size())

				value, err := pq.Peek()
				require.Nil(t, err)
				require.Equal(t, testCase.peeks[index], value)
			}

			for _, top := range testCase.pops {
				cur_size := pq.Size()
				value, err := pq.Peek()
				require.Nil(t, err)
				require.Equal(t, top, value)

				value, err = pq.Pop()
				require.Nil(t, err)
				require.Equal(t, top, value)
				require.Equal(t, cur_size-1, pq.Size())
			}

			require.Equal(t, 0, pq.Size())

			_, err := pq.Peek()
			require.Equal(t, err.Code, errs.InvalidOperation)

			_, err = pq.Pop()
			require.Equal(t, err.Code, errs.InvalidOperation)
		})
	}

}

func TestPriorityQuery_String(t *testing.T) {
	testCases := []struct {
		name          string
		compare       Comparator[string]
		initialValues []string
		inserts       []string
		peeks         []string
		pops          []string
	}{
		{
			name:          "min priority queue",
			compare:       CompareAsc[string],
			initialValues: []string{},
			inserts: []string{
				"1",
				"3",
				"5",
				"2",
				"4",
				"0",
			},
			peeks: []string{"1", "1", "1", "1", "1", "0"},
			pops:  []string{"0", "1", "2", "3", "4", "5"},
		},
		{
			name:          "max priority queue",
			compare:       CompareDesc[string],
			initialValues: []string{},
			inserts: []string{
				"1",
				"3",
				"5",
				"2",
				"4",
				"0",
			},
			peeks: []string{"1", "3", "5", "5", "5", "5"},
			pops:  []string{"5", "4", "3", "2", "1", "0"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			pq := NewPriorityQueue(
				testCase.compare,
				testCase.initialValues,
			)

			require.Equal(t, len(testCase.initialValues), pq.Size())

			for index, insert := range testCase.inserts {
				pq.Insert(insert)
				require.Equal(t, index+1, pq.Size())

				value, _ := pq.Peek()
				require.Equal(t, testCase.peeks[index], value)
			}

			for index, top := range testCase.pops {
				value, err := pq.Peek()
				require.Nil(t, err)
				require.Equal(t, top, value)

				value, err = pq.Pop()
				require.Nil(t, err)
				require.Equal(t, top, value)
				require.Equal(t, len(testCase.pops)-index-1, pq.Size())
			}

			require.Equal(t, 0, pq.Size())

			_, err := pq.Peek()
			require.Equal(t, err.Code, errs.InvalidOperation)

			_, err = pq.Pop()
			require.Equal(t, err.Code, errs.InvalidOperation)
		})
	}

}

func TestPriorityQuery_Date(t *testing.T) {
	testCases := []struct {
		name          string
		compare       func(value1, value2 time.Time) Comparison
		initialValues []time.Time
		inserts       []time.Time
		peeks         []time.Time
		pops          []time.Time
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
			initialValues: []time.Time{},
			inserts: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			peeks: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			pops: []time.Time{
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
			initialValues: []time.Time{},
			inserts: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			peeks: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
				time.Date(2020, 1, 1, 0, 0, 0, 5, time.UTC),
			},
			pops: []time.Time{
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
				testCase.initialValues,
			)

			require.Equal(t, len(testCase.initialValues), pq.Size())

			for index, insert := range testCase.inserts {
				pq.Insert(insert)
				require.Equal(t, index+1, pq.Size())

				value, err := pq.Peek()
				require.Nil(t, err)
				require.Equal(t, testCase.peeks[index], value)
			}

			for index, top := range testCase.pops {
				value, err := pq.Peek()
				require.Nil(t, err)
				require.Equal(t, top, value)

				value, err = pq.Pop()
				require.Nil(t, err)
				require.Equal(t, top, value)
				require.Equal(t, len(testCase.pops)-index-1, pq.Size())
			}

			require.Equal(t, 0, pq.Size())

			_, err := pq.Peek()
			require.Equal(t, err.Code, errs.InvalidOperation)

			_, err = pq.Pop()
			require.Equal(t, err.Code, errs.InvalidOperation)
		})
	}
}
