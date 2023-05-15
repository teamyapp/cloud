package algo

import "golang.org/x/exp/constraints"

type PriorityQueueItem[Value any, Priority constraints.Ordered] struct {
	value    Value
	priority Priority
}

type PriorityQueue[Value any, Priority constraints.Ordered] struct {
	items      []PriorityQueueItem[Value, Priority]
	comparator func(priority1, priority2 Priority) bool
}

func (pq *PriorityQueue[Value, Priority]) shiftUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if pq.hasHigherPriority(pq.items[index], pq.items[parent]) {
			pq.items[index], pq.items[parent] = pq.items[parent], pq.items[index]
			index = parent
		} else {
			break
		}
	}
}

func (pq *PriorityQueue[Value, Priority]) hasHigherPriority(item1, item2 PriorityQueueItem[Value, Priority]) bool {
	return pq.comparator(item1.priority, item2.priority)
}

func (pq *PriorityQueue[Value, Priority]) shiftDown(index int) {
	left := index*2 + 1
	right := index*2 + 2

	largest := index

	if left < pq.Size() && pq.hasHigherPriority(pq.items[left], pq.items[largest]) {
		largest = left
	}

	if right < pq.Size() && pq.hasHigherPriority(pq.items[right], pq.items[largest]) {
		largest = right
	}

	if largest != index {
		pq.items[index], pq.items[largest] = pq.items[largest], pq.items[index]
		pq.shiftDown(largest)
	}
}

func (pq *PriorityQueue[Value, Priority]) Size() int {
	return len(pq.items)
}

func (pq *PriorityQueue[Value, Priority]) Push(value Value, priority Priority) {
	pq.items = append(pq.items, PriorityQueueItem[Value, Priority]{
		value:    value,
		priority: priority,
	})

	pq.shiftUp(pq.Size() - 1)
}

func (pq *PriorityQueue[Value, Priority]) Pop() *Value {
	if pq.Size() == 0 {
		return nil
	}

	value := pq.items[0].value
	pq.items[0] = pq.items[pq.Size()-1]
	pq.items = pq.items[:pq.Size()-1]
	pq.shiftDown(0)
	return &value
}

func (pq *PriorityQueue[Value, Priority]) Peek() *Value {
	if pq.Size() == 0 {
		return nil
	}

	return &pq.items[0].value
}

func NewPriorityQueue[Value any, Priority constraints.Ordered](
	comparator func(priority1, priority2 Priority) bool,
) *PriorityQueue[Value, Priority] {
	return &PriorityQueue[Value, Priority]{
		items:      make([]PriorityQueueItem[Value, Priority], 0),
		comparator: comparator,
	}
}
