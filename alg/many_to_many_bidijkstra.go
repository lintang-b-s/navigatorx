package alg

import (
	"container/heap"
	"fmt"
	"lintang/navigatorx/util"
	"math"
)

type SPSingleResultResult struct {
	Source int32
	Dest   int32
	Paths  []CHNode2
	Dist   float64
	Eta    float64
}

func (ch *ContractedGraph) callBidirectionalDijkstra(spMap []int32) SPSingleResultResult {
	var path []CHNode2
	var eta float64
	var dist float64
	path, eta, dist = ch.ShortestPathBiDijkstra(spMap[0], spMap[1])

	return SPSingleResultResult{spMap[0], spMap[1], path, dist, eta}
}

func (ch *ContractedGraph) ShortestPathManyToManyBiDijkstraWorkers(from []int32, to []int32) map[int32]map[int32]SPSingleResultResult {
	spPair := [][]int32{}
	for i := 0; i < len(from); i++ {
		for j := 0; j < len(to); j++ {

			spPair = append(spPair, []int32{from[i], to[j]})
		}
	}
	workers := NewWorkerPool[[]int32, SPSingleResultResult](3, len(spPair))

	for i := 0; i < len(spPair); i++ {
		workers.AddJob(spPair[i])
	}
	close(workers.jobQueue)

	spMap := make(map[int32]map[int32]SPSingleResultResult)

	workers.Start(ch.callBidirectionalDijkstra)
	workers.Wait()

	for i := 0; i < len(spPair); i++ {
		spMap[spPair[i][0]] = make(map[int32]SPSingleResultResult)
	}

	for curr := range workers.CollectResults() {

		spMap[curr.Source][curr.Dest] = curr
	}

	return spMap
}

// https://github.com/RoutingKit/RoutingKit/blob/master/src/contraction_hierarchy.cpp [void pinned_run(...)]
// lemot... 5 detik buat 3 source & 4 destination
func (ch *ContractedGraph) ShortestPathManyToManyBiDijkstra(from int32, to []int32) ([][]CHNode2, []float64, []float64) {
	forwQ := &priorityQueue[CHNode2]{}
	backwQ := &priorityQueue[CHNode2]{}

	containInBackwQ := make(map[int32]bool)

	df := make(map[int32]float64)
	df[from] = 0.0

	nmf := nodeMapCHBiDijkstra{}

	heap.Init(forwQ)
	heap.Init(backwQ)

	pqLookup := make(map[interface{}]*priorityQueueNode[CHNode2])
	fromNode := nmf.getCHDJ2(ch.ContractedNodes[from])
	fromNode.rank = 0

	if fromNode == nil {
		fmt.Println("fromNode is nil")
	}

	heap.Push(forwQ, fromNode)
	wasForwardPushed := make(map[int32]bool)

	cameFromf := make(map[int32]cameFromPair)
	cameFromf[from] = cameFromPair{EdgeCH{}, -1}

	target_list := make([]int32, len(to))
	for i := 0; i < len(to); i++ {
		t := to[i]
		target_list[i] = t
		_, ok := containInBackwQ[t]
		if !ok {
			target := nmf.getCHDJ2(ch.ContractedNodes[t])
			heap.Push(backwQ, target)
			containInBackwQ[t] = true
		}
	}

	selectList := []int32{}
	for {
		if backwQ.Len() == 0 {
			break
		}
		node := heap.Pop(backwQ).(*priorityQueueNode[CHNode2])
		selectList = append(selectList, node.item.IDx)

		for _, arc := range ch.ContractedFirstInEdge[node.item.IDx] {
			edge := ch.ContractedInEdges[arc]

			_, ok := containInBackwQ[edge.ToNodeIDX]
			if !ok {
				toN := nmf.getCHDJ2(ch.ContractedNodes[edge.ToNodeIDX])
				heap.Push(backwQ, toN)
				containInBackwQ[edge.ToNodeIDX] = true
			}
		}

	}

	selectList = util.ReverseG(selectList)

	frontier := forwQ
	for {
		if frontier.Len() == 0 {
			break
		}

		node := heap.Pop(frontier).(*priorityQueueNode[CHNode2])

		for _, arc := range ch.ContractedFirstOutEdge[node.item.IDx] {
			edge := ch.ContractedOutEdges[arc]
			toNIDx := edge.ToNodeIDX
			cost := edge.Weight

			newCost := cost + node.rank
			val, ok := df[toNIDx]
			if !ok || newCost < val {
				df[toNIDx] = newCost
				neighborNode := nmf.getCHDJ2(ch.ContractedNodes[toNIDx])
				neighborNode.rank = newCost
				heap.Push(frontier, neighborNode)
				pqLookup[neighborNode.item] = neighborNode

				cameFromf[toNIDx] = cameFromPair{edge, node.item.IDx}
			}
		}

	}

	for i := range selectList {
		x := selectList[i]
		dist := math.Inf(1)
		pred := int32(-1)
		if val, ok := wasForwardPushed[x]; ok && val {
			dist = df[x]
		}

		// harus tanpa compare node order buat many to many (udah coba pake node order salah hasilnya)
		for _, arc := range ch.ContractedFirstInEdge[x] {
			edge := ch.ContractedInEdges[arc]
			y := edge.ToNodeIDX

			newDist := df[y] + edge.Weight
			if newDist < dist {
				dist = newDist
				pred = arc
			}

		}
		nn := cameFromf[x].NodeIDx
		if pred != -1 {
			df[x] = dist

			cameFromf[x] = cameFromPair{ch.ContractedInEdges[pred], nn}
		} else if dist == math.Inf(1) {
			df[x] = math.Inf(1)
			cameFromf[x] = cameFromPair{EdgeCH{}, -1}
		}
	}

	targetDist := make([]float64, len(target_list))
	targetPath := make([][]CHNode2, len(target_list))
	targetEta := make([]float64, len(target_list))
	for i := range target_list {
		t := target_list[i]
		targetEta[i] = 0
		ok := true
		for ok && t != -1 {
			if cameFromf[t].Edge.IsShortcut {
				if t != target_list[i] {
					ch.unpackBackward(cameFromf[t].Edge, &targetPath[i], &targetEta[i], &targetDist[i])
				} else {
					//  target_list[i] pake backward InEdges
					ch.unpackForward(cameFromf[t].Edge, &targetPath[i], &targetEta[i], &targetDist[i])
				}
			} else {
				_, ok = cameFromf[t]
				if ok && cameFromf[t].NodeIDx != -1 && ch.ContractedNodes[cameFromf[t].NodeIDx].TrafficLight {
					targetEta[i] += 3.0
				}
				targetPath[i] = append(targetPath[i], ch.ContractedNodes[t])

				targetDist[i] += cameFromf[t].Edge.Dist
				targetEta[i] += cameFromf[t].Edge.Weight
			}
			_, ok = cameFromf[t]
			t = cameFromf[t].NodeIDx
		}
		targetDist[i] = targetDist[i] / 1000
	}

	return targetPath, targetDist, targetEta
}
