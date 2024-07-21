package alg

import (
	"container/heap"
	"fmt"
	"math"
)

type cameFromPair struct {
	Edge    EdgePair
	NodeIDx int32
}

/*
referensi:
- https://github.com/jgrapht/jgrapht/blob/master/jgrapht-core/src/main/java/org/jgrapht/alg/shortestpath/ContractionHierarchyBidirectionalDijkstra.java
- https://github.com/navjindervirdee/Advanced-Shortest-Paths-Algorithms/blob/master/Contraction%20Hierarchies/DistPreprocessSmall.java

*/

func (ch *ContractedGraph) ShortestPathBiDijkstra(from, to int32) ([]CHNode, float64, float64) {
	forwQ := &priorityQueueDijkstra{}
	backQ := &priorityQueueDijkstra{}
	df := make(map[int32]float64)
	db := make(map[int32]float64)
	df[from] = 0.0
	db[to] = 0.0

	nmf := nodeMapCHDijkstra{}
	nmb := nodeMapCHDijkstra{}

	heap.Init(forwQ)
	heap.Init(backQ)

	fromNode := nmf.getCHDJ(ch.OrigGraph[from])
	fromNode.rank = 0
	toNode := nmb.getCHDJ(ch.OrigGraph[to])
	toNode.rank = 0
	heap.Push(forwQ, fromNode)
	heap.Push(backQ, toNode)

	estimate := math.MaxFloat64

	bestCommonVertex := int32(0)

	cameFromf := make(map[int32]cameFromPair)
	cameFromf[from] = cameFromPair{EdgePair{}, -1}

	cameFromb := make(map[int32]cameFromPair)
	cameFromb[to] = cameFromPair{EdgePair{}, -1}

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
		if ff[0].rank >= estimate {
			if turnF {
				frontFinished = true
			} else {
				backFinished = true
			}
		} else {
			node := heap.Pop(frontier).(*dijkstraNode)
			if node.rank > estimate {
				break
			}
			if turnF {
				for _, edge := range ch.OrigGraph[node.CHNode.IDx].OutEdges {
					toNIDx := edge.ToNodeIDX
					cost := edge.Weight
					if ch.OrigGraph[node.CHNode.IDx].orderPos < ch.OrigGraph[toNIDx].orderPos {
						// upward graph
						newCost := cost + df[node.CHNode.IDx]
						_, ok := df[toNIDx]
						if !ok || newCost < df[toNIDx] {
							df[toNIDx] = newCost
							neighborNode := nmf.getCHDJ(ch.OrigGraph[toNIDx])
							neighborNode.rank = newCost
							heap.Push(frontier, neighborNode)

							cameFromf[toNIDx] = cameFromPair{edge, node.CHNode.IDx}
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
				for _, edge := range ch.OrigGraph[node.CHNode.IDx].InEdges {
					toNIDx := edge.ToNodeIDX
					cost := edge.Weight
					if ch.OrigGraph[node.CHNode.IDx].orderPos < ch.OrigGraph[toNIDx].orderPos {
						// upward graph
						newCost := cost + db[node.CHNode.IDx]
						_, ok := db[toNIDx]
						if !ok || newCost < db[toNIDx] {
							db[toNIDx] = newCost

							neighborNode := nmf.getCHDJ(ch.OrigGraph[toNIDx])
							neighborNode.rank = newCost
							heap.Push(frontier, neighborNode)

							cameFromb[toNIDx] = cameFromPair{edge, node.CHNode.IDx}
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
		return []CHNode{}, -1, -1
	}
	// estimate dari bidirectional dijkstra pake shortcut edge jadi lebih cepet eta nya & gak akurat
	path, eta, dist := ch.createPath(bestCommonVertex, from, to, cameFromf, cameFromb)
	return path, eta, dist
}

func (ch *ContractedGraph) createPath(commonVertex int32, from, to int32,
	cameFromf, cameFromb map[int32]cameFromPair) ([]CHNode, float64, float64) {

	// edges := []EdgePair{}
	fPath := []CHNode{}
	eta := 0.0
	dist := 0.0
	v := commonVertex
	if ch.OrigGraph[v].TrafficLight {
		eta += 2.0
	}
	for v != -1 {

		if cameFromf[v].Edge.IsShortcut {

			ch.unpackBackward(cameFromf[v].Edge, &fPath, &eta, &dist)
		} else {

			if cameFromf[v].NodeIDx != -1 && ch.OrigGraph[cameFromf[v].NodeIDx].TrafficLight {
				eta += 2.0
			}
			eta += cameFromf[v].Edge.Weight
			dist += cameFromf[v].Edge.Dist
			fPath = append(fPath, ch.OrigGraph[v])
		}
		v = cameFromf[v].NodeIDx

	}

	bPath := []CHNode{}
	v = commonVertex
	for v != -1 {

		if cameFromb[v].Edge.IsShortcut {

			ch.unpackForward(cameFromb[v].Edge, &bPath, &eta, &dist)

		} else {

			if cameFromb[v].NodeIDx != -1 && ch.OrigGraph[cameFromb[v].NodeIDx].TrafficLight {
				eta += 2.0
			}
			eta += cameFromb[v].Edge.Weight
			dist += cameFromb[v].Edge.Dist
			bPath = append(bPath, ch.OrigGraph[v])
		}
		v = cameFromb[v].NodeIDx

	}

	fPath = reverseCH(fPath)[:len(fPath)-1]
	path := []CHNode{}
	path = append(path, fPath...)
	path = append(path, bPath...)
	tf := 0
	for _, p := range path {
		if p.TrafficLight {
			tf++
		}

	}
	fmt.Println(tf)
	return path, eta, dist / 1000
}

// buat forward dijkstra
// dari common vertex ke source vertex
func (ch *ContractedGraph) unpackBackward(edge EdgePair, path *[]CHNode, eta, dist *float64) {
	if !edge.IsShortcut {
		if ch.OrigGraph[edge.ToNodeIDX].TrafficLight {
			*eta += 2.0
		}
		*eta += edge.Weight
		*dist += edge.Dist
		*path = append(*path, ch.OrigGraph[edge.ToNodeIDX])
	} else {
		ch.unpackBackward(*edge.RemovedEdgeTwo, path, eta, dist)
		ch.unpackBackward(*edge.RemovedEdgeOne, path, eta, dist)
	}
}

// dari common vertex ke target vertex
func (ch *ContractedGraph) unpackForward(edge EdgePair, path *[]CHNode, eta, dist *float64) {
	if !edge.IsShortcut {
		if ch.OrigGraph[edge.ToNodeIDX].TrafficLight {
			*eta += 2.0
		}
		*eta += edge.Weight
		*dist += edge.Dist
		*path = append(*path, ch.OrigGraph[edge.ToNodeIDX])
	} else {
		ch.unpackForward(*edge.RemovedEdgeOne, path, eta, dist)
		ch.unpackForward(*edge.RemovedEdgeTwo, path, eta, dist)

	}
}
