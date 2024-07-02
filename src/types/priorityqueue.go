package types

import (
	"container/heap"
	"fmt"
)

type Item struct {
	Value    string
	Priority int
	Index    int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq PriorityQueue) Peek() any {
	n := len(pq)
	item := pq[n-1]
	return item
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Update(item *Item, Priority int) {
	item.Priority = Priority
	heap.Fix(pq, item.Index)

	for i := range *pq {
		fmt.Printf("%v, ", *(*pq)[i])
	}
	fmt.Println()
}

func (pq PriorityQueue) Get(Index int) *Item {
	return pq[Index]
}
