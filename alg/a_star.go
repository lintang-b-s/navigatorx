package alg

import (
	"container/heap"

	"github.com/twpayne/go-polyline"
)



type nodeMapCH map[int32]*priorityQueueNode[CHNode]

func (nm nodeMapCH) getCH(p CHNode) *priorityQueueNode[CHNode] {
	n, ok := nm[p.IDx]

	if !ok {
		n = &priorityQueueNode[CHNode]{item: p}

		nm[p.IDx] = n
	}
	return n
}

// https://theory.stanford.edu/~amitp/GameProgramming/ImplementationNotes.html
func (ch *ContractedGraph) AStarCH(from, to int32) (pathN []CHNode, path string, eta float64, found bool, dist float64) {
	nm := nodeMapCH{}
	nq := &priorityQueue[CHNode2]{}
	heap.Init(nq)
	fromNode := nm.getCH(ch.AStarGraph[from])
	fromNode.rank = 0
	heap.Push(nq, fromNode)
	costSoFar := make(map[int32]float64)
	costSoFar[ch.AStarGraph[from].IDx] = 0.0
	distSoFar := make(map[int32]float64)
	distSoFar[ch.AStarGraph[from].IDx] = 0.0

	cameFrom := make(map[int32]*CHNode)
	cameFrom[ch.AStarGraph[from].IDx] = nil

	for {
		if nq.Len() == 0 {
			return
		}

		current := heap.Pop(nq).(*priorityQueueNode[CHNode])
		if current == nm.getCH(ch.AStarGraph[to]) {
			s := ""
			etaTraffic := 0.0

			path := []CHNode{}
			curr := current
			for curr.rank != 0 {
				if curr.item.TrafficLight {
					etaTraffic += 2.0
				}
				path = append(path, curr.item)
				curr = nm.getCH(ch.AStarGraph[cameFrom[curr.item.IDx].IDx])
			}
			path = append(path, ch.AStarGraph[from])
			path = reverseG(path)
			coords := make([][]float64, 0)

			pathN := []CHNode{}
			for _, p := range path {
				pathN = append(pathN, p)
				coords = append(coords, []float64{p.Lat, p.Lon})
			}
			s = string(polyline.EncodeCoords(coords))

			return pathN, s, costSoFar[current.item.IDx] + etaTraffic, true, distSoFar[current.item.IDx] / 1000
		}

		for _, neighbor := range current.item.OutEdges {
			newCost := costSoFar[current.item.IDx] + neighbor.Weight
			dist := distSoFar[current.item.IDx] + neighbor.Dist
			neighborP := ch.AStarGraph[neighbor.ToNodeIDX]
			neighborNode := nm.getCH(neighborP)
			_, ok := costSoFar[neighborP.IDx]
			if !ok || newCost < costSoFar[neighborP.IDx] {
				costSoFar[neighborP.IDx] = newCost
				distSoFar[neighborP.IDx] = dist
				cameFrom[neighborP.IDx] = &ch.AStarGraph[current.item.IDx]
				priority := newCost + neighborP.PathEstimatedCostETA(ch.AStarGraph[to])
				neighborNode.rank = priority
				heap.Push(nq, neighborNode)
			}
		}
	}
}
