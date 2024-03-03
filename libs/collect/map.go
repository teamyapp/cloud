package collect

func Map[From any, To any](
	items []From,
	transform func(fromItem From, index int) To,
) []To {
	newItems := make([]To, len(items))
	for index, item := range items {
		newItems[index] = transform(item, index)
	}

	return newItems
}

func MapWithErr[From any, To any, Err any](
	items []From,
	transform func(fromItem From, index int) (To, *Err),
) ([]To, *Err) {
	newItems := make([]To, 0)
	for index, item := range items {
		newItem, err := transform(item, index)
		if err != nil {
			return nil, err
		}

		newItems = append(newItems, newItem)
	}

	return newItems, nil
}
