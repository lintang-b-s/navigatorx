package alg

type priorityQueueDijkstra []*dijkstraNode

func (pq priorityQueueDijkstra) Len() int {
	return len(pq)
}

func (pq priorityQueueDijkstra) Less(i, j int) bool {
	return pq[i].rank < pq[j].rank
}

func (pq priorityQueueDijkstra) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueueDijkstra) Push(x interface{}) {
	n := len(*pq)
	no := x.(*dijkstraNode)
	no.index = n
	*pq = append(*pq, no)
}

func (pq *priorityQueueDijkstra) Pop() interface{} {
	old := *pq
	n := len(old)
	no := old[n-1]
	no.index = -1
	*pq = old[0 : n-1]
	return no
}
