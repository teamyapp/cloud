package algo

type PriorityQueueItem[Value any] struct {
	value Value
}

type PriorityQueue[Value any] struct {
	items   []Value
	compare Comparator
}

func (pq *PriorityQueue[Value]) shiftUp(index int) {
	for index > 0 {
		parentIndex := (index - 1) / 2
		if parentIndex  == 0 || pq.hasHigherPriority(pq.items[parentIndex], pq.items[index]) {
			break
		}

		pq.items[index], pq.items[parentIndex] = pq.items[parentIndex], pq.items[index]
		index = parentIndex
	}
}

func (pq *PriorityQueue[Value]) hasHigherPriority(item1, item2 PriorityQueueItem[Value]) bool {
	return pq.compare(item1.value, item2.value) == SmallerThan
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

func leftChildIndex(index int) int {
	return index*2 + 1
}

func rightChildIndex(index int) int {
	return index*2 + 2
}

func (pq *PriorityQueue[Value]) Size() int {
	return len(pq.items)
}

func (pq *PriorityQueue[Value]) Push(value Value) {
	pq.items = append(pq.items, value)
	pq.shiftUp(pq.Size() - 1)
}

func (pq *PriorityQueue[Value]) Pop() (Value, error) {
	if pq.Size() == 0 {
		return nil
	}

	root := pq.items[0]
	pq.items[0] = pq.items[pq.Size()-1]
	pq.items = pq.items[:pq.Size()-1]
	pq.shiftDown(0)
	return root, nil
}

func (pq *PriorityQueue[Value]) Peek() (Value, error) {
	if pq.Size() == 0 {
		return nil
	}

	return pq.items[0].value, nil
}

func NewPriorityQueue[Value any](
	comparator func(value1, value2 Value) int,
) *PriorityQueue[Value] {
	return &PriorityQueue[Value]{
		items:   make([]Value, 0),
		compare: compare,
	}
}
