package delta

func DetectMapDelta[Key comparable, Value any, ValueDelta any](
	oldMap map[Key]Value,
	newMap map[Key]Value,
	detectValueDelta func(oldValue Value, newValue Value) Delta[ValueDelta],
	toDelta func(status Status, value Value) ValueDelta,
) Delta[map[Key]KeyValueDelta[ValueDelta]] {
	status := UnchangedStatus
	deltaMap := make(map[Key]KeyValueDelta[ValueDelta])
	for newKey, newValue := range newMap {
		if oldValue, ok := oldMap[newKey]; ok {
			valueDelta := detectValueDelta(oldValue, newValue)
			deltaMap[newKey] = KeyValueDelta[ValueDelta]{
				KeyStatus:   UnchangedStatus,
				ValueStatus: valueDelta.Status,
				Value:       valueDelta.Value,
			}

			if valueDelta.Status != UnchangedStatus {
				status = UpdatedStatus
			}
		} else {
			keyValueDelta := KeyValueDelta[ValueDelta]{
				KeyStatus:   AddedStatus,
				ValueStatus: AddedStatus,
				Value:       toDelta(AddedStatus, newValue),
			}
			deltaMap[newKey] = keyValueDelta
			status = UpdatedStatus
		}
	}

	for oldKey, oldValue := range oldMap {
		if _, ok := newMap[oldKey]; !ok {
			keyValueDelta := KeyValueDelta[ValueDelta]{
				KeyStatus:   RemovedStatus,
				ValueStatus: RemovedStatus,
				Value:       toDelta(RemovedStatus, oldValue),
			}
			deltaMap[oldKey] = keyValueDelta
			status = UpdatedStatus
		}
	}

	return Delta[map[Key]KeyValueDelta[ValueDelta]]{
		Status: status,
		Value:  deltaMap,
	}
}

func ToMapDelta[Key comparable, Value any, ValueDelta any](
	status Status,
	input map[Key]Value,
	toValueDelta func(status Status, value Value) ValueDelta,
) map[Key]KeyValueDelta[ValueDelta] {
	output := make(map[Key]KeyValueDelta[ValueDelta])
	for key, val := range input {
		output[key] = KeyValueDelta[ValueDelta]{
			KeyStatus:   status,
			ValueStatus: status,
			Value:       toValueDelta(status, val),
		}
	}

	return output
}
