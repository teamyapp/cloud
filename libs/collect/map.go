package collect

func Map[From any, To any, Err any](items []From, transform func(fromItem From, index int) (To, *Err)) ([]To, *Err) {
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
