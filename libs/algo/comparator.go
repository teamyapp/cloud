package algo

import "golang.org/x/exp/constraints"

const (
	Equal       int = 0
	GreaterThan     = 1
	SmallerThan     = 2
)

func DefaultCompareDesc[Value constraints.Ordered](value1 Value, value2 Value) int {
	if value1 == value2 {
		return Equal
	}

	if value1 > value2 {
		return GreaterThan
	}

	return SmallerThan
}

func DefaultCompareAsc[Value constraints.Ordered](value1 Value, value2 Value) int {
	if value1 == value2 {
		return Equal
	}

	if value1 > value2 {
		return SmallerThan
	}

	return GreaterThan
}
