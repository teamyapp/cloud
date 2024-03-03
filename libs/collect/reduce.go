package collect

func Reduce[Item any, Value any](
	items []Item,
	accumulate func(accumulatedValue Value, item Item) Value,
	initialValue Value,
) Value {
	accumulatedValue := initialValue
	for _, item := range items {
		accumulatedValue = accumulate(accumulatedValue, item)
	}

	return accumulatedValue
}

func ReduceWithErr[Item any, Value any, Err any](
	items []Item,
	accumulate func(accumulatedValue Value, item Item) (Value, *Err),
	initialValue Value,
) (Value, *Err) {
	accumulatedValue := initialValue
	for _, item := range items {
		newAccumulatedValue, err := accumulate(accumulatedValue, item)
		if err != nil {
			return newAccumulatedValue, err
		}

		accumulatedValue = newAccumulatedValue
	}

	return accumulatedValue, nil
}
