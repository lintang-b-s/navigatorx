package alg


type priorityQueueNodeOrdering []*pqCHNode

func (pq priorityQueueNodeOrdering) Len() int {
	return len(pq)
}

func (pq priorityQueueNodeOrdering) Less(i, j int) bool {
	return pq[i].rank < pq[j].rank
}

func (pq priorityQueueNodeOrdering) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueueNodeOrdering) Push(x interface{}) {
	n := len(*pq)
	no := x.(*pqCHNode)
	no.index = n
	*pq = append(*pq, no)
}

func (pq *priorityQueueNodeOrdering) Pop() interface{} {
	old := *pq
	n := len(old)
	no := old[n-1]
	no.index = -1
	*pq = old[0 : n-1]
	return no
}
