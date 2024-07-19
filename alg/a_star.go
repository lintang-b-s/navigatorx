package alg

import (
	"container/heap"

	"github.com/twpayne/go-polyline"
)

type astarNodeCH struct {
	chNode *CHNode
	rank   float64
	parent *astarNodeCH
	index  int
}

type nodeMapCH map[*CHNode]*astarNodeCH

func (nm nodeMapCH) getCH(p *CHNode) *astarNodeCH {
	n, ok := nm[p]

	if !ok {
		n = &astarNodeCH{chNode: p}

		nm[p] = n
	}
	return n
}

// https://theory.stanford.edu/~amitp/GameProgramming/ImplementationNotes.html
func AStarCH(from, to int32) (pathN []CHNode, path string, eta float64, found bool, dist float64) {
	nm := nodeMapCH{}
	nq := &priorityQueueCH{}
	heap.Init(nq)
	fromNode := nm.getCH(CHGraph.AStarGraph[from])
	fromNode.rank = 0
	heap.Push(nq, fromNode)
	costSoFar := make(map[*CHNode]float64)
	costSoFar[CHGraph.AStarGraph[from]] = 0.0
	distSoFar := make(map[*CHNode]float64)
	distSoFar[CHGraph.AStarGraph[from]] = 0.0

	cameFrom := make(map[*CHNode]*CHNode)
	cameFrom[CHGraph.AStarGraph[from]] = nil

	for {
		if nq.Len() == 0 {
			return
		}

		current := heap.Pop(nq).(*astarNodeCH)
		if current == nm.getCH(CHGraph.AStarGraph[to]) {
			s := ""
			etaTraffic := 0.0

			path := []CHNode{}
			curr := current
			for curr.rank != 0 {
				if curr.chNode.TrafficLight {
					etaTraffic += 2.0
				}
				path = append(path, *curr.chNode)
				curr = nm.getCH(cameFrom[curr.chNode])
			}
			path = append(path, *CHGraph.AStarGraph[from])
			path = reverseCH(path)
			coords := make([][]float64, 0)

			pathN := []CHNode{}
			for _, p := range path {
				pathN = append(pathN, p)
				coords = append(coords, []float64{p.Lat, p.Lon})
			}
			s = string(polyline.EncodeCoords(coords))

			return pathN, s, costSoFar[current.chNode] + etaTraffic, true, distSoFar[current.chNode] / 1000
		}

		for _, neighbor := range current.chNode.OutEdges {
			newCost := costSoFar[current.chNode] + neighbor.Weight
			dist := distSoFar[current.chNode] + neighbor.Dist
			neighborP := CHGraph.AStarGraph[neighbor.ToNodeIDX]
			neighborNode := nm.getCH(neighborP)
			_, ok := costSoFar[neighborP]
			if !ok || newCost < costSoFar[neighborP] {
				costSoFar[neighborP] = newCost
				distSoFar[neighborP] = dist
				cameFrom[neighborP] = current.chNode
				priority := newCost + neighborP.PathEstimatedCostETA(*CHGraph.AStarGraph[to])
				neighborNode.rank = priority
				heap.Push(nq, neighborNode)
			}
		}
	}
}
