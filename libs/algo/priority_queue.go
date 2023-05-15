package algo

type PriorityQueueItem[Value any] struct {
	value Value
}

type PriorityQueue[Value any] struct {
	items   []PriorityQueueItem[Value]
	compare func(value1, value2 Value) int
}

func (pq *PriorityQueue[Value]) shiftUp(index int) {
	for index > 0 {
		parentIndex := (index - 1) / 2
		if pq.hasHigherPriority(pq.items[parentIndex], pq.items[index]) {
			break
		}

		pq.items[index], pq.items[parentIndex] = pq.items[parentIndex], pq.items[index]
		index = parentIndex
	}
}

func (pq *PriorityQueue[Value]) hasHigherPriority(item1, item2 PriorityQueueItem[Value]) bool {
	return pq.compare(item1.value, item2.value) == GreaterThan
}

func (pq *PriorityQueue[Value]) shiftDown(index int) {
	for pq.leftChildIndex(index) < pq.Size() {
		leftChildIndex := pq.leftChildIndex(index)
		rightChildIndex := pq.rightChildIndex(index)

		largerChildIndex := leftChildIndex

		if rightChildIndex < pq.Size() && pq.hasHigherPriority(pq.items[rightChildIndex], pq.items[leftChildIndex]) {
			largerChildIndex = rightChildIndex
		}

		if pq.hasHigherPriority(pq.items[index], pq.items[largerChildIndex]) {
			break
		}

		pq.items[index], pq.items[largerChildIndex] = pq.items[largerChildIndex], pq.items[index]
	}
}

func (pq *PriorityQueue[Value]) leftChildIndex(index int) int {
	return index*2 + 1
}

func (pq *PriorityQueue[Value]) rightChildIndex(index int) int {
	return index*2 + 2
}

func (pq *PriorityQueue[Value]) Size() int {
	return len(pq.items)
}

func (pq *PriorityQueue[Value]) Push(value Value) {
	pq.items = append(pq.items, PriorityQueueItem[Value]{
		value: value,
	})
	pq.shiftUp(pq.Size() - 1)
}

func (pq *PriorityQueue[Value]) Pop() *Value {
	if pq.Size() == 0 {
		return nil
	}

	value := pq.items[0].value
	pq.items[0] = pq.items[pq.Size()-1]
	pq.items = pq.items[:pq.Size()-1]
	pq.shiftDown(0)
	return &value
}

func (pq *PriorityQueue[Value]) Peek() *Value {
	if pq.Size() == 0 {
		return nil
	}

	return &pq.items[0].value
}

func NewPriorityQueue[Value any](
	compare func(value1, value2 Value) int,
) *PriorityQueue[Value] {
	return &PriorityQueue[Value]{
		items:   make([]PriorityQueueItem[Value], 0),
		compare: compare,
	}
}
