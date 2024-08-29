package alg

import (
	"container/heap"
	"fmt"
	"lintang/navigatorx/util"
	"math"
)

type cameFromPair struct {
	Edge    EdgeCH
	NodeIDx int32
}

type nodeMapCHBiDijkstra map[int32]*priorityQueueNode[CHNode2]

func (nm nodeMapCHBiDijkstra) getCHDJ2(p CHNode2) *priorityQueueNode[CHNode2] {
	n, ok := nm[p.IDx]

	if !ok {
		var bidijkstraNode priorityQueueNode[CHNode2]
		bidijkstraNode.item = p
		n = &bidijkstraNode

		nm[p.IDx] = n
	}
	return n
}

/*
referensi:
- https://github.com/jgrapht/jgrapht/blob/master/jgrapht-core/src/main/java/org/jgrapht/alg/shortestpath/ContractionHierarchyBidirectionalDijkstra.java
- https://github.com/navjindervirdee/Advanced-Shortest-Paths-Algorithms/blob/master/Contraction%20Hierarchies/DistPreprocessSmall.java

*/

func (ch *ContractedGraph) ShortestPathBiDijkstra(from, to int32) ([]CHNode2, float64, float64) {
	forwQ := &priorityQueue[CHNode2]{}
	backQ := &priorityQueue[CHNode2]{}
	df := make(map[int32]float64)
	db := make(map[int32]float64)
	df[from] = 0.0
	db[to] = 0.0

	nmf := nodeMapCHBiDijkstra{}
	nmb := nodeMapCHBiDijkstra{}

	heap.Init(forwQ)
	heap.Init(backQ)

	fromNode := nmf.getCHDJ2(ch.ContractedNodes[from])
	fromNode.rank = 0
	toNode := nmb.getCHDJ2(ch.ContractedNodes[to])
	toNode.rank = 0

	if fromNode == nil {
		fmt.Println("fromNode is nil")
	}
	if toNode == nil {
		fmt.Println("toNode is nil")
	}
	heap.Push(forwQ, fromNode)
	heap.Push(backQ, toNode)

	estimate := math.MaxFloat64

	bestCommonVertex := int32(0)

	cameFromf := make(map[int32]cameFromPair)
	cameFromf[from] = cameFromPair{EdgeCH{}, -1}

	cameFromb := make(map[int32]cameFromPair)
	cameFromb[to] = cameFromPair{EdgeCH{}, -1}

	frontFinished := false
	backFinished := false

	frontier := forwQ
	otherFrontier := backQ
	turnF := true
	for {
		if frontier.Len() == 0 {
			frontFinished = true
		}
		if otherFrontier.Len() == 0 {
			backFinished = true
		}

		if frontFinished && backFinished {
			break
		}

		ff := *frontier
		if len(ff) == 0 {
			return []CHNode2{}, -1, -1
		}
		if ff[0].rank >= estimate {
			if turnF {
				frontFinished = true
			} else {
				backFinished = true
			}
		} else {
			node := heap.Pop(frontier).(*priorityQueueNode[CHNode2])
			if node.rank > estimate {
				break
			}
			if turnF {

				for _, arc := range ch.ContractedFirstOutEdge[node.item.IDx] {
					edge := ch.ContractedOutEdges[arc]
					toNIDx := edge.ToNodeIDX
					cost := edge.Weight
					if ch.ContractedNodes[node.item.IDx].OrderPos < ch.ContractedNodes[toNIDx].OrderPos {
						// upward graph
						newCost := cost + df[node.item.IDx]
						_, ok := df[toNIDx]
						if !ok || newCost < df[toNIDx] {
							df[toNIDx] = newCost
							neighborNode := nmf.getCHDJ2(ch.ContractedNodes[toNIDx])
							neighborNode.rank = newCost
							heap.Push(frontier, neighborNode)

							cameFromf[toNIDx] = cameFromPair{edge, node.item.IDx}
						}

						_, ok = db[toNIDx]
						if ok {
							pathDistance := newCost + db[toNIDx]
							if pathDistance < estimate {
								estimate = pathDistance
								bestCommonVertex = edge.ToNodeIDX

							}
						}
					}
				}

			} else {

				for _, arc := range ch.ContractedFirstInEdge[node.item.IDx] {

					edge := ch.ContractedInEdges[arc]
					toNIDx := edge.ToNodeIDX
					cost := edge.Weight
					if ch.ContractedNodes[node.item.IDx].OrderPos < ch.ContractedNodes[toNIDx].OrderPos {
						// upward graph
						newCost := cost + db[node.item.IDx]
						_, ok := db[toNIDx]
						if !ok || newCost < db[toNIDx] {
							db[toNIDx] = newCost

							neighborNode := nmb.getCHDJ2(ch.ContractedNodes[toNIDx])
							neighborNode.rank = newCost
							heap.Push(frontier, neighborNode)

							cameFromb[toNIDx] = cameFromPair{edge, node.item.IDx}
						}

						_, ok = df[toNIDx]
						if ok {
							pathDistance := newCost + df[toNIDx]
							if pathDistance < estimate {
								estimate = pathDistance
								bestCommonVertex = edge.ToNodeIDX

							}
						}
					}
				}

			}

		}

		otherFinished := false

		if turnF {
			if backFinished {
				otherFinished = true
			}
		} else {
			if frontFinished {
				otherFinished = true
			}

		}
		if !otherFinished {
			tmpFrontier := frontier
			frontier = otherFrontier
			otherFrontier = tmpFrontier
			turnF = !turnF
		}
	}

	if estimate == math.MaxFloat64 {
		return []CHNode2{}, -1, -1
	}
	// estimate dari bidirectional dijkstra pake shortcut edge jadi lebih cepet eta nya & gak akurat
	path, eta, dist := ch.createPath(bestCommonVertex, from, to, cameFromf, cameFromb)
	return path, eta, dist
}

func (ch *ContractedGraph) createPath(commonVertex int32, from, to int32,
	cameFromf, cameFromb map[int32]cameFromPair) ([]CHNode2, float64, float64) {

	// edges := []EdgePair{}
	fPath := []CHNode2{}
	eta := 0.0
	dist := 0.0
	v := commonVertex
	if ch.ContractedNodes[v].TrafficLight {
		eta += 3.0
	}
	ok := true
	for ok && v != -1 {

		if cameFromf[v].Edge.IsShortcut {

			ch.unpackBackward(cameFromf[v].Edge, &fPath, &eta, &dist)
		} else {

			if cameFromf[v].NodeIDx != -1 && ch.ContractedNodes[cameFromf[v].NodeIDx].TrafficLight {
				eta += 3.0
			}
			eta += cameFromf[v].Edge.Weight
			dist += cameFromf[v].Edge.Dist
			fPath = append(fPath, ch.ContractedNodes[v])
		}
		_, ok = cameFromf[v]
		v = cameFromf[v].NodeIDx

	}

	bPath := []CHNode2{}
	v = commonVertex
	ok = true
	for ok && v != -1 {

		if cameFromb[v].Edge.IsShortcut {

			ch.unpackForward(cameFromb[v].Edge, &bPath, &eta, &dist)

		} else {

			if cameFromb[v].NodeIDx != -1 && ch.ContractedNodes[cameFromb[v].NodeIDx].TrafficLight {
				eta += 3.0
			}
			eta += cameFromb[v].Edge.Weight
			dist += cameFromb[v].Edge.Dist
			bPath = append(bPath, ch.ContractedNodes[v])
		}
		_, ok = cameFromb[v]
		v = cameFromb[v].NodeIDx
	}

	fPath = util.ReverseG(fPath)[:len(fPath)-1]
	path := []CHNode2{}
	path = append(path, fPath...)
	path = append(path, bPath...)
	tf := 0
	for _, p := range path {
		if p.TrafficLight {
			tf++
		}

	}
	// fmt.Println(tf)
	return path, eta, dist / 1000
}

// buat forward dijkstra
// dari common vertex ke source vertex
func (ch *ContractedGraph) unpackBackward(edge EdgeCH, path *[]CHNode2, eta *float64, dist *float64) {
	if !edge.IsShortcut {
		if ch.ContractedNodes[edge.ToNodeIDX].TrafficLight {
			*eta += 3.0
		}
		*eta += edge.Weight
		*dist += edge.Dist
		*path = append(*path, ch.ContractedNodes[edge.ToNodeIDX])
	} else {
		ch.unpackBackward(ch.ContractedOutEdges[edge.RemovedEdgeTwo], path, eta, dist)
		ch.unpackBackward(ch.ContractedOutEdges[edge.RemovedEdgeOne], path, eta, dist)
	}
}

// dari common vertex ke target vertex
func (ch *ContractedGraph) unpackForward(edge EdgeCH, path *[]CHNode2, eta *float64, dist *float64) {
	if !edge.IsShortcut {
		if ch.ContractedNodes[edge.ToNodeIDX].TrafficLight {
			*eta += 3.0
		}
		*eta += edge.Weight
		*dist += edge.Dist
		*path = append(*path, ch.ContractedNodes[edge.ToNodeIDX])
	} else {
		ch.unpackForward(ch.ContractedInEdges[edge.RemovedEdgeOne], path, eta, dist)
		ch.unpackForward(ch.ContractedInEdges[edge.RemovedEdgeTwo], path, eta, dist)

	}
}
