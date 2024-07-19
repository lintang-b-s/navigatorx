package alg

import (
	"container/heap"
	"math"
)

type dijkstraNode struct {
	CHNode *CHNode
	rank   float64
	index  int
}

type nodeMapCHDijkstra map[*CHNode]*dijkstraNode

func (nm nodeMapCHDijkstra) getCHDJ(p *CHNode) *dijkstraNode {
	n, ok := nm[p]

	if !ok {
		n = &dijkstraNode{CHNode: p}

		nm[p] = n
	}
	return n
}

func dijkstraWitnessSearch(fromNodeIDx, targetNodeIDx int32, ignoreNodeIDx int32,
	acceptedWeight float64, maxSettledNodes int, pMax float64, contracted []bool) float64 {

	visited := make(map[int32]bool)
	cost := make(map[int32]float64)

	nm := nodeMapCHDijkstra{}
	nq := &priorityQueueDijkstra{}
	heap.Init(nq)
	fromNode := nm.getCHDJ(CHGraph.OrigGraph[fromNodeIDx])
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

		curr := heap.Pop(nq).(*dijkstraNode)
		// if contracted[curr.CHNode.IDx] {
		// 	// continue
		// }
		if curr == nm.getCHDJ(CHGraph.OrigGraph[targetNodeIDx]) {
			return cost[curr.CHNode.IDx]
		}

		if curr.rank > pMax {
			out, ok := cost[targetNodeIDx]
			if ok {
				return out
			}
			return math.MaxFloat64
		}

		visited[curr.CHNode.IDx] = true
		for _, neighbor := range curr.CHNode.OutEdges {
			if visited[neighbor.ToNodeIDX] || neighbor.ToNodeIDX == ignoreNodeIDx ||
				contracted[neighbor.ToNodeIDX] {
				continue
			}

			neighborP := CHGraph.OrigGraph[neighbor.ToNodeIDX]
			neighborNode := nm.getCHDJ(neighborP)
			newCost := cost[curr.CHNode.IDx] + neighbor.Weight
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
