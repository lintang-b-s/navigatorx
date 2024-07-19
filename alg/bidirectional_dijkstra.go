package alg

import (
	"container/heap"
	"fmt"
	"math"
)

type cameFromPair struct {
	NodeIDx int32
	Edge    EdgePair
}

/*
referensi: 
- https://github.com/jgrapht/jgrapht/blob/master/jgrapht-core/src/main/java/org/jgrapht/alg/shortestpath/ContractionHierarchyBidirectionalDijkstra.java
- https://github.com/navjindervirdee/Advanced-Shortest-Paths-Algorithms/blob/master/Contraction%20Hierarchies/DistPreprocessSmall.java

*/

func ShortestPathBiDijkstra(from, to int32) ([]CHNode, float64, float64) {
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

	fromNode := nmf.getCHDJ(CHGraph.OrigGraph[from])
	fromNode.rank = 0
	toNode := nmb.getCHDJ(CHGraph.OrigGraph[to])
	toNode.rank = 0
	heap.Push(forwQ, fromNode)
	heap.Push(backQ, toNode)

	estimate := math.MaxFloat64

	bestCommonVertex := int32(0)

	cameFromf := make(map[int32]cameFromPair)
	cameFromf[from] = cameFromPair{-1, EdgePair{}}

	cameFromb := make(map[int32]cameFromPair)
	cameFromb[to] = cameFromPair{-1, EdgePair{}}

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
			// nIDx := node.CHNode.IDx
			if node.rank > estimate {
				break
			}
			if turnF {
				for _, edge := range CHGraph.OrigGraph[node.CHNode.IDx].OutEdges {
					toNIDx := edge.ToNodeIDX
					cost := edge.Weight
					if CHGraph.OrigGraph[node.CHNode.IDx].orderPos < CHGraph.OrigGraph[toNIDx].orderPos {
						// upward graph
						newCost := cost + df[node.CHNode.IDx]
						_, ok := df[toNIDx]
						if !ok || newCost < df[toNIDx] {
							df[toNIDx] = newCost
							// etaf[toNIDx] = etaf[node.CHNode.IDx] + edge.Weight
							// distf[toNIDx] = distf[node.CHNode.IDx] + edge.Dist
							neighborNode := nmf.getCHDJ(CHGraph.OrigGraph[toNIDx])
							neighborNode.rank = newCost
							heap.Push(frontier, neighborNode)

							cameFromf[toNIDx] = cameFromPair{node.CHNode.IDx, edge}
						}

						_, ok = db[toNIDx]
						if ok {
							pathDistance := newCost + db[toNIDx]
							if pathDistance < estimate {
								estimate = pathDistance
								bestCommonVertex = edge.ToNodeIDX
								// eta = etaf[toNIDx] + etab[toNIDx]
								// dist = distf[toNIDx] + distb[toNIDx]
							}
						}
					}
				}

				// _, ok := db[nIDx]
				// if ok {
				// 	pathDistance := df[nIDx] + db[nIDx]
				// 	if pathDistance < estimate {
				// 		estimate = pathDistance
				// 		bestCommonVertex = nIDx
				// 	}
				// }

			} else {
				for _, edge := range CHGraph.OrigGraph[node.CHNode.IDx].InEdges {
					toNIDx := edge.ToNodeIDX
					cost := edge.Weight
					if CHGraph.OrigGraph[node.CHNode.IDx].orderPos < CHGraph.OrigGraph[toNIDx].orderPos {
						// downward graph
						newCost := cost + db[node.CHNode.IDx]
						_, ok := db[toNIDx]
						if !ok || newCost < db[toNIDx] {
							db[toNIDx] = newCost
							// etab[toNIDx] = etab[node.CHNode.IDx] + edge.Weight
							// distb[toNIDx] = distb[node.CHNode.IDx] + edge.Dist
							neighborNode := nmf.getCHDJ(CHGraph.OrigGraph[toNIDx])
							neighborNode.rank = newCost
							heap.Push(frontier, neighborNode)

							cameFromb[toNIDx] = cameFromPair{node.CHNode.IDx, edge}
						}

						_, ok = df[toNIDx]
						if ok {
							pathDistance := newCost + df[toNIDx]
							if pathDistance < estimate {
								estimate = pathDistance
								bestCommonVertex = edge.ToNodeIDX
								// eta = etaf[toNIDx] + etab[toNIDx]
								// dist = distf[toNIDx] + distb[toNIDx]
							}
						}
					}
				}

				// _, ok := df[nIDx]
				// if ok {
				// 	pathDistance := db[nIDx] + df[nIDx]
				// 	if pathDistance < estimate {
				// 		estimate = pathDistance
				// 		bestCommonVertex = nIDx
				// 	}
				// }
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
	path, eta, dist := createPath(bestCommonVertex, from, to, cameFromf, cameFromb)
	return path, eta, dist
}

func createPath(commonVertex int32, from, to int32,
	cameFromf, cameFromb map[int32]cameFromPair) ([]CHNode, float64, float64) {

	// edges := []EdgePair{}
	fPath := []CHNode{}
	eta := 0.0
	dist := 0.0
	v := commonVertex
	if CHGraph.OrigGraph[v].TrafficLight {
		eta += 2.0
	}
	for v != -1 {

		if cameFromf[v].Edge.IsShortcut {

			unpackBackward(cameFromf[v].Edge, &fPath, &eta, &dist)
		} else {

			if cameFromf[v].NodeIDx != -1 && CHGraph.OrigGraph[cameFromf[v].NodeIDx].TrafficLight {
				eta += 2.0
			}
			eta += cameFromf[v].Edge.Weight
			dist += cameFromf[v].Edge.Dist
			fPath = append(fPath, *CHGraph.OrigGraph[v])
		}
		v = cameFromf[v].NodeIDx

	}

	bPath := []CHNode{}
	v = commonVertex
	for v != -1 {

		if cameFromb[v].Edge.IsShortcut {

			unpackForward(cameFromb[v].Edge, &bPath, &eta, &dist)

		} else {

			if cameFromb[v].NodeIDx != -1 && CHGraph.OrigGraph[cameFromb[v].NodeIDx].TrafficLight {
				eta += 2.0
			}
			eta += cameFromb[v].Edge.Weight
			dist += cameFromb[v].Edge.Dist
			bPath = append(bPath, *CHGraph.OrigGraph[v])
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
func unpackBackward(edge EdgePair, path *[]CHNode, eta, dist *float64) {
	if !edge.IsShortcut {
		if CHGraph.OrigGraph[edge.ToNodeIDX].TrafficLight {
			*eta += 2.0
		}
		*eta += edge.Weight
		*dist += edge.Dist
		*path = append(*path, *CHGraph.OrigGraph[edge.ToNodeIDX])
	} else {
		unpackBackward(*edge.RemovedEdgeTwo, path, eta, dist)
		unpackBackward(*edge.RemovedEdgeOne, path, eta, dist)
	}
}

// dari common vertex ke target vertex
func unpackForward(edge EdgePair, path *[]CHNode, eta, dist *float64) {
	if !edge.IsShortcut {
		if CHGraph.OrigGraph[edge.ToNodeIDX].TrafficLight {
			*eta += 2.0
		}
		*eta += edge.Weight
		*dist += edge.Dist
		*path = append(*path, *CHGraph.OrigGraph[edge.ToNodeIDX])
	} else {
		unpackForward(*edge.RemovedEdgeOne, path, eta, dist)
		unpackForward(*edge.RemovedEdgeTwo, path, eta, dist)

	}
}

// salah gara gara orderpos
// func ShortestPathBiDijkstra(from, to int32) ([]CHNode, float64) {
// 	forwQ := &priorityQueueDijkstra{}
// 	backQ := &priorityQueueDijkstra{}
// 	df := make(map[int32]float64)
// 	db := make(map[int32]float64)
// 	df[from] = 0.0
// 	db[to] = 0.0

// 	sf := make(map[int32]bool) // visited forward search
// 	sb := make(map[int32]bool) // visited backward search

// 	nmf := nodeMapCHDijkstra{}
// 	nmb := nodeMapCHDijkstra{}

// 	heap.Init(forwQ)
// 	heap.Init(backQ)

// 	fromNode := nmf.getCHDJ(CHGraph.OrigGraph[from])
// 	toNode := nmb.getCHDJ(CHGraph.OrigGraph[to])
// 	heap.Push(forwQ, fromNode)
// 	heap.Push(backQ, toNode)

// 	estimate := math.MaxFloat64

// 	bestCommonVertex := int32(0)

// 	cameFromf := make(map[int32]int32)
// 	cameFromf[from] = -1

// 	cameFromb := make(map[int32]int32)
// 	cameFromb[to] = -1

// 	for forwQ.Len() != 0 || backQ.Len() != 0 {
// 		if forwQ.Len() != 0 {
// 			v1 := heap.Pop(forwQ).(*dijkstraNode)

// 			if df[v1.CHNode.IDx] <= estimate {
// 				sf[v1.CHNode.IDx] = true
// 				for _, edge := range CHGraph.OrigGraph[v1.CHNode.IDx].Edges {
// 					toNIDx := edge.ToNodeIDX
// 					cost := edge.Weight
// 					if CHGraph.OrigGraph[v1.CHNode.IDx].orderPos < CHGraph.OrigGraph[toNIDx].orderPos {
// 						// upward graph
// 						newCost := cost + df[v1.CHNode.IDx]
// 						_, ok := df[toNIDx]
// 						if !ok ||  newCost < df[toNIDx] {
// 							df[toNIDx] = newCost
// 							heap.Push(forwQ, nmf.getCHDJ(CHGraph.OrigGraph[toNIDx]))
// 							bestCommonVertex = edge.ToNodeIDX
// 							cameFromf[toNIDx] = v1.CHNode.IDx
// 						}
// 					}
// 				}
// 			}

// 			if sb[v1.CHNode.IDx] {
// 				if df[v1.CHNode.IDx]+db[v1.CHNode.IDx] < estimate {
// 					estimate = df[v1.CHNode.IDx] + db[v1.CHNode.IDx]
// 				}
// 			}
// 		}

// 		if backQ.Len() != 0 {
// 			v2 := heap.Pop(backQ).(*dijkstraNode)
// 			if db[v2.CHNode.IDx] <= estimate {
// 				sb[v2.CHNode.IDx] = true
// 				for _, edge := range CHGraph.RevGraph[v2.CHNode.IDx].Edges {
// 					toNIDx := edge.ToNodeIDX
// 					cost := edge.Weight
// 					if CHGraph.RevGraph[v2.CHNode.IDx].orderPos > CHGraph.RevGraph[toNIDx].orderPos {
// 						// downward graph
// 						newCost := cost + db[v2.CHNode.IDx]
// 						_, ok := db[toNIDx]
// 						if !ok || newCost < db[toNIDx] {
// 							db[toNIDx] = newCost
// 							heap.Push(backQ, nmb.getCHDJ(CHGraph.RevGraph[toNIDx]))
// 							bestCommonVertex = edge.ToNodeIDX
// 							cameFromb[toNIDx] = v2.CHNode.IDx
// 						}
// 					}
// 				}
// 			}

// 			if sf[v2.CHNode.IDx] {
// 				if db[v2.CHNode.IDx]+df[v2.CHNode.IDx] < estimate {
// 					estimate = db[v2.CHNode.IDx] + df[v2.CHNode.IDx]
// 				}
// 			}
// 		}
// 	}
// 	if estimate == math.MaxFloat64 {
// 		return []CHNode{}, -1
// 	}

// 	path := createPath(bestCommonVertex, from, to, cameFromf, cameFromb)
// 	return path, estimate
// }
