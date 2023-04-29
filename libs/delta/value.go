package delta

type Delta[Value any] struct {
	Status Status
	Value  Value
}

type KeyValueDelta[Value any] struct {
	KeyStatus   Status
	ValueStatus Status
	Value       Value
}

func DetectValueDelta[Value comparable](oldValue Value, newValue Value) Delta[Value] {
	status := UnchangedStatus
	if oldValue != newValue {
		status = UpdatedStatus
	}

	return Delta[Value]{
		Status: status,
		Value:  newValue,
	}
}

func ToValueDelta[Value any](_ Status, value Value) Value {
	return value
}
