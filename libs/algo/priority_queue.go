package algo

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type PriorityQueue[Value any] struct {
	items   []Value
	compare Comparator[Value]
}

func (pq *PriorityQueue[Value]) Size() int {
	return len(pq.items)
}

func (pq *PriorityQueue[Value]) Insert(value Value) {
	pq.items = append(pq.items, value)
	pq.shiftUp(pq.Size() - 1)
}

func (pq *PriorityQueue[Value]) Pop() (Value, *errs.Error) {
	if pq.Size() == 0 {
		return *new(Value), errs.NewError(errs.InvalidOperation, "empty priority queue")
	}

	root := pq.items[0]
	pq.items[0] = pq.items[pq.Size()-1]
	pq.items = pq.items[:pq.Size()-1]
	pq.shiftDown(0)
	return root, nil
}

func (pq *PriorityQueue[Value]) Peek() (Value, *errs.Error) {
	if pq.Size() == 0 {
		return *new(Value), errs.NewError(errs.InvalidOperation, "empty priority queue")
	}

	return pq.items[0], nil
}

func (pq *PriorityQueue[Value]) Remove(value *Value) (Value, *errs.Error) {
	if pq.Size() == 0 {
		return *new(Value), errs.NewError(errs.InvalidOperation, "priority queue is empty")
	}

	for index, item := range pq.items {
		if &item == value {
			pq.items[index] = pq.items[pq.Size()-1]
			pq.items = pq.items[:pq.Size()-1]
			pq.shiftDown(index)
			return item, nil
		}
	}

	return *new(Value), errs.NewError(errs.InvalidOperation, "value not found")
}

func (pq *PriorityQueue[Value]) Items() []Value {
	return pq.items
}

func (pq *PriorityQueue[Value]) heapify() {
	for index := pq.Size()/2 - 1; index >= 0; index-- {
		pq.shiftDown(index)
	}
}

func leftChildIndex(index int) int {
	return index*2 + 1
}

func rightChildIndex(index int) int {
	return index*2 + 2
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

func (pq *PriorityQueue[Value]) shiftDown(index int) {
	for leftChildIndex(index) < pq.Size() {
		leftChildIdx := leftChildIndex(index)
		rightChildIdx := rightChildIndex(index)

		largerChildIdx := leftChildIdx
		if rightChildIdx < pq.Size() && pq.hasHigherPriority(pq.items[rightChildIdx], pq.items[leftChildIdx]) {
			largerChildIdx = rightChildIdx
		}

		if pq.hasHigherPriority(pq.items[index], pq.items[largerChildIdx]) {
			break
		}

		pq.items[index], pq.items[largerChildIdx] = pq.items[largerChildIdx], pq.items[index]
		index = largerChildIdx
	}
}

func (pq *PriorityQueue[Value]) hasHigherPriority(item1, item2 Value) bool {
	return pq.compare(item1, item2) == SmallerThan
}

func NewPriorityQueue[Value any](
	comparator Comparator[Value],
	items []Value,
) *PriorityQueue[Value] {
	pq := &PriorityQueue[Value]{
		items:   items,
		compare: comparator,
	}

	pq.heapify()
	return pq
}
