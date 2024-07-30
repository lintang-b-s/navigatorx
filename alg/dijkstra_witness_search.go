package alg

import (
	"container/heap"
	"math"
)

type dijkstraNode struct {
	rank   float32
	index  int
	CHNode CHNode2
}

type nodeMapCHDijkstra map[int32]*dijkstraNode

func (nm nodeMapCHDijkstra) getCHDJ(p CHNode2) *dijkstraNode {
	n, ok := nm[p.IDx]

	if !ok {
		n = &dijkstraNode{CHNode: p}

		nm[p.IDx] = n
	}
	return n
}

func (ch *ContractedGraph) dijkstraWitnessSearch(fromNodeIDx, targetNodeIDx int32, ignoreNodeIDx int32,
	acceptedWeight float32, maxSettledNodes int, pMax float32, contracted []bool) float32 {

	visited := make(map[int32]bool)
	cost := make(map[int32]float32)

	nm := nodeMapCHDijkstra{}
	nq := &priorityQueueDijkstra{}
	heap.Init(nq)
	fromNode := nm.getCHDJ(ch.ContractedNodes[fromNodeIDx])
	fromNode.rank = 0
	heap.Push(nq, fromNode)

	cost[fromNodeIDx] = 0.0
	settledNodes := 0
	for {
		pq := *nq
		if nq.Len() == 0 || pq[0].rank > acceptedWeight || settledNodes >= maxSettledNodes {
			return math.MaxFloat32
		}

		_, ok := cost[targetNodeIDx]
		if ok && cost[targetNodeIDx] <= acceptedWeight {
			return cost[targetNodeIDx]
		}

		curr := heap.Pop(nq).(*dijkstraNode)
		if contracted[curr.CHNode.IDx] {
			continue
		}
		if curr == nm.getCHDJ(ch.ContractedNodes[targetNodeIDx]) {
			return cost[curr.CHNode.IDx]
		}

		if curr.rank > pMax {
			out, ok := cost[targetNodeIDx]
			if ok {
				return out
			}
			return math.MaxFloat32
		}

		visited[curr.CHNode.IDx] = true
		for _, outIDx := range ch.ContractedFirstOutEdge[curr.CHNode.IDx] {
			neighbor := ch.ContractedOutEdges[outIDx]
			if visited[neighbor.ToNodeIDX] || neighbor.ToNodeIDX == ignoreNodeIDx ||
				contracted[neighbor.ToNodeIDX] {
				continue
			}

			neighborP := ch.ContractedNodes[neighbor.ToNodeIDX]
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
