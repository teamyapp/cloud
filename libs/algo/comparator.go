package algo

type Comparison int

const (
	Equal       Comparison = 0
	GreaterThan Comparison = 1
	SmallerThan Comparison = 2
)

type Comparator[Value any] func(value1 Value, value2 Value) Comparison

type Comparable interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64 |
		string
}

func CompareAsc[Value Comparable](value1 Value, value2 Value) Comparison {
	if value1 == value2 {
		return Equal
	}

	if value1 > value2 {
		return GreaterThan
	}

	return SmallerThan
}

var _ Comparator[int] = CompareAsc[int]

func CompareDesc[Value Comparable](value1 Value, value2 Value) Comparison {
	if value1 == value2 {
		return Equal
	}

	if value1 > value2 {
		return SmallerThan
	}

	return GreaterThan
}

var _ Comparator[int] = CompareDesc[int]
