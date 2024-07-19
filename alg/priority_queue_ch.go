package alg

type priorityQueueCH []*astarNodeCH

func (pq priorityQueueCH) Len() int {
	return len(pq)
}

func (pq priorityQueueCH) Less(i, j int) bool {
	return pq[i].rank < pq[j].rank
}

func (pq priorityQueueCH) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueueCH) Push(x interface{}) {
	n := len(*pq)
	no := x.(*astarNodeCH)
	no.index = n
	*pq = append(*pq, no)
}

func (pq *priorityQueueCH) Pop() interface{} {
	old := *pq
	n := len(old)
	no := old[n-1]
	no.index = -1
	*pq = old[0 : n-1]
	return no
}
