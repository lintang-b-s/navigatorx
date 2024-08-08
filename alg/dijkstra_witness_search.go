package alg

import (
	"container/heap"
	"math"
)


func (ch *ContractedGraph) dijkstraWitnessSearch(fromNodeIDx, targetNodeIDx int32, ignoreNodeIDx int32,
	acceptedWeight float64, maxSettledNodes int, pMax float64, contracted []bool) float64 {

	visited := make(map[int32]bool)
	cost := make(map[int32]float64)

	nm := nodeMapCHBiDijkstra{}
	nq := &priorityQueue[CHNode2]{}
	heap.Init(nq)
	fromNode := nm.getCHDJ2(ch.ContractedNodes[fromNodeIDx])
	fromNode.rank = 0
	heap.Push(nq, fromNode)

	cost[fromNodeIDx] = 0.0
	settledNodes := 0
	for {
		pq := *nq
		if nq.Len() == 0 || pq[0].rank > acceptedWeight || settledNodes >= maxSettledNodes {
			return math.MaxFloat64
		}

		_, ok := cost[targetNodeIDx]
		if ok && cost[targetNodeIDx] <= acceptedWeight {
			return cost[targetNodeIDx]
		}

		curr := heap.Pop(nq).(*priorityQueueNode[CHNode2])
		if contracted[curr.item.IDx] {
			continue
		}
		if curr == nm.getCHDJ2(ch.ContractedNodes[targetNodeIDx]) {
			return cost[curr.item.IDx]
		}

		if curr.rank > pMax {
			out, ok := cost[targetNodeIDx]
			if ok {
				return out
			}
			return math.MaxFloat64
		}

		visited[curr.item.IDx] = true
		for _, outIDx := range ch.ContractedFirstOutEdge[curr.item.IDx] {
			neighbor := ch.ContractedOutEdges[outIDx]
			if visited[neighbor.ToNodeIDX] || neighbor.ToNodeIDX == ignoreNodeIDx ||
				contracted[neighbor.ToNodeIDX] {
				continue
			}

			neighborP := ch.ContractedNodes[neighbor.ToNodeIDX]
			neighborNode := nm.getCHDJ2(neighborP)
			newCost := cost[curr.item.IDx] + neighbor.Weight
			_, ok := cost[neighbor.ToNodeIDX]
			if !ok || newCost < cost[neighbor.ToNodeIDX] {
				cost[neighbor.ToNodeIDX] = newCost

				neighborNode.rank = newCost
				heap.Push(nq, neighborNode)
			}
		}

		settledNodes++
	}
}
