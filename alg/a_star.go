package alg

import (
	"container/heap"

	"github.com/twpayne/go-polyline"
)

type astarNodeCH struct {
	rank   float32
	chNode CHNode
	parent *astarNodeCH
	index  int
}

type nodeMapCH map[int32]*astarNodeCH

func (nm nodeMapCH) getCH(p CHNode) *astarNodeCH {
	n, ok := nm[p.IDx]

	if !ok {
		n = &astarNodeCH{chNode: p}

		nm[p.IDx] = n
	}
	return n
}

// https://theory.stanford.edu/~amitp/GameProgramming/ImplementationNotes.html
func (ch *ContractedGraph) AStarCH(from, to int32) (pathN []CHNode, path string, eta float64, found bool, dist float64) {
	nm := nodeMapCH{}
	nq := &priorityQueueCH{}
	heap.Init(nq)
	fromNode := nm.getCH(ch.AStarGraph[from])
	fromNode.rank = 0
	heap.Push(nq, fromNode)
	costSoFar := make(map[int32]float32)
	costSoFar[ch.AStarGraph[from].IDx] = 0.0
	distSoFar := make(map[int32]float32)
	distSoFar[ch.AStarGraph[from].IDx] = 0.0

	cameFrom := make(map[int32]*CHNode)
	cameFrom[ch.AStarGraph[from].IDx] = nil

	for {
		if nq.Len() == 0 {
			return
		}

		current := heap.Pop(nq).(*astarNodeCH)
		if current == nm.getCH(ch.AStarGraph[to]) {
			s := ""
			etaTraffic := 0.0

			path := []CHNode{}
			curr := current
			for curr.rank != 0 {
				if curr.chNode.TrafficLight {
					etaTraffic += 2.0
				}
				path = append(path, curr.chNode)
				curr = nm.getCH(ch.AStarGraph[cameFrom[curr.chNode.IDx].IDx])
			}
			path = append(path, ch.AStarGraph[from])
			path = reverseCH(path)
			coords := make([][]float64, 0)

			pathN := []CHNode{}
			for _, p := range path {
				pathN = append(pathN, p)
				coords = append(coords, []float64{float64(p.Lat), float64(p.Lon)})
			}
			s = string(polyline.EncodeCoords(coords))

			return pathN, s, float64(costSoFar[current.chNode.IDx]) + etaTraffic, true, float64(distSoFar[current.chNode.IDx] / 1000)
		}

		for _, neighbor := range current.chNode.OutEdges {
			newCost := costSoFar[current.chNode.IDx] + neighbor.Weight
			dist := distSoFar[current.chNode.IDx] + neighbor.Dist
			neighborP := ch.AStarGraph[neighbor.ToNodeIDX]
			neighborNode := nm.getCH(neighborP)
			_, ok := costSoFar[neighborP.IDx]
			if !ok || newCost < costSoFar[neighborP.IDx] {
				costSoFar[neighborP.IDx] = newCost
				distSoFar[neighborP.IDx] = dist
				cameFrom[neighborP.IDx] = &ch.AStarGraph[current.chNode.IDx]
				priority := newCost + float32(neighborP.PathEstimatedCostETA(ch.AStarGraph[to]))
				neighborNode.rank = priority
				heap.Push(nq, neighborNode)
			}
		}
	}
}
