package collect

func Filter[Item any, Err any](
	items []Item,
	match func(item Item) (bool, *Err),
) ([]Item, *Err) {
	matchedItems := make([]Item, 0)
	for _, item := range items {
		ok, err := match(item)
		if err != nil {
			return nil, err
		}

		if ok {
			matchedItems = append(matchedItems, item)
		}
	}

	return matchedItems, nil
}
