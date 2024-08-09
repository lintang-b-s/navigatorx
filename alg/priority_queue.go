package alg

type Item interface {
	CHNode2 | CHNode | int32
}
type priorityQueueNode[T Item] struct {
	rank  float64
	index int
	item  T
}

type priorityQueue[T Item] []*priorityQueueNode[T]

func (pq priorityQueue[Item]) Len() int {
	return len(pq)
}

func (pq priorityQueue[Item]) Less(i, j int) bool {
	return pq[i].rank < pq[j].rank
}

func (pq priorityQueue[Item]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue[Item]) Push(x interface{}) {
	n := len(*pq)
	no := x.(*priorityQueueNode[Item])
	no.index = n
	*pq = append(*pq, no)
}

func (pq *priorityQueue[Item]) Pop() interface{} {
	old := *pq
	n := len(old)
	no := old[n-1]
	no.index = -1
	*pq = old[0 : n-1]
	return no
}

