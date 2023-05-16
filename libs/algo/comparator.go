package algo

type Comparison int
const (
	Equal        Comparison = 0
	GreaterThan  Comparison = 1
	SmallerThan  Comparison = 2
)
type Comparator [Value any] func(value1 Value, value2 Value) Comparison
func CompareAsc[Value comparable](value1 Value, value2 Value) Comparison {
	if value1 == value2 {
		return Equal
	}

	if value1 > value2 {
		return GreaterThan
	}

	return SmallerThan
}

var _ Comparator = CompareAsc

func CompareDesc[Value comparable](value1 Value, value2 Value) Comparison {
	if value1 == value2 {
		return Equal
	}

	if value1 > value2 {
		return SmallerThan
	}

	return GreaterThan
}

var _ Comparator = CompareDesc
