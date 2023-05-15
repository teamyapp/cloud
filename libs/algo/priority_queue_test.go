package algo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQuery_Int(t *testing.T) {
	testCases := []struct {
		name    string
		compare func(value1, value2 int) int
		inserts []int
		peaks   []int
		tops    []int
	}{
		{
			name:    "min priority queue",
			compare: DefaultCompareAsc[int],
			inserts: []int{1, 3, 5, 2, 4, 0},
			peaks:   []int{1, 1, 1, 1, 1, 0},
			tops:    []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:    "max priority queue",
			compare: DefaultCompareDesc[int],
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
				testCase.compare,
			)

			assert.Equal(t, 0, pq.Size())

			for index, insert := range testCase.inserts {
				pq.Push(insert)
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
		name    string
		compare func(value1, value2 string) int
		inserts []string
		peaks   []string
		tops    []string
	}{
		{
			name:    "min priority queue",
			compare: DefaultCompareAsc[string],
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
			name:    "max priority queue",
			compare: DefaultCompareDesc[string],
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
			)

			assert.Equal(t, 0, pq.Size())

			for index, insert := range testCase.inserts {
				pq.Push(insert)
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

func TestPriorityQuery_Date(t *testing.T) {

	testCases := []struct {
		name    string
		compare func(value1, value2 time.Time) int
		inserts []time.Time
		peaks   []time.Time
		tops    []time.Time
	}{
		{
			name: "min priority queue",
			compare: func(time1 time.Time, time2 time.Time) int {
				if time1.Before(time2) {
					return GreaterThan
				}

				if time1.After(time2) {
					return SmallerThan
				}

				return Equal
			},
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
			compare: func(time1 time.Time, time2 time.Time) int {
				if time1.Before(time2) {
					return SmallerThan
				}

				if time1.After(time2) {
					return GreaterThan
				}

				return Equal
			},
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
			)

			assert.Equal(t, 0, pq.Size())

			for index, insert := range testCase.inserts {
				pq.Push(insert)
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
