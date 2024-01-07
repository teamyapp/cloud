package matcher

import (
	"strings"
)

type Filter[Item any] func(item Item) bool

// Logical filters

func And[Item any](filter1 Filter[Item], filter2 Filter[Item]) Filter[Item] {
	return func(item Item) bool {
		return filter1(item) && filter2(item)
	}
}

func Or[Item any](filter1 Filter[Item], filter2 Filter[Item]) Filter[Item] {
	return func(item Item) bool {
		return filter1(item) || filter2(item)
	}
}

func Not[Item any](filter Filter[Item]) Filter[Item] {
	return func(item Item) bool {
		return !filter(item)
	}
}

// Comparison filters
type Selector[Item any] func(item Item) interface{}

func EqualTo[Item any, Value Equatable](selector Selector[Item], target Value) Filter[Item] {
	return func(item Item) bool {
		return selector(item).(Value) == target
	}
}

func Contains[Item any](selector Selector[Item], target string) Filter[Item] {
	return func(item Item) bool {
		return strings.Contains(selector(item).(string), target)
	}
}

func GreaterThan[Item any, Value Comparable](selector Selector[Item], target Value) Filter[Item] {
	return func(item Item) bool {
		return selector(item).(Value) > target
	}
}

func GreaterThanOrEqualTo[Item any, Value Comparable](selector Selector[Item], target Value) Filter[Item] {
	return func(item Item) bool {
		return selector(item).(Value) >= target
	}
}

func LessThan[Item any, Value Comparable](selector Selector[Item], target Value) Filter[Item] {
	return func(item Item) bool {
		return selector(item).(Value) < target
	}
}

func LessThanOrEqualTo[Item any, Value Comparable](selector Selector[Item], target Value) Filter[Item] {
	return func(item Item) bool {
		return selector(item).(Value) <= target
	}
}
