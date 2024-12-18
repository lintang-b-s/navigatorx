package contractor

import (
	"math"
)

/*
	dijkstraWitnessSearch
	misal kita kontraksi node v (ignoreNodeIDx), kita harus cari shortest path dari node u ke w yang meng ignore node v, dimana u adalah salah satu node yang terhubung ke v dan edge (u,v) \in E, dan w adalah  salah satu node yang terhubung dari v dan edge (v,w) \in E.
	search dihentikan jika current visited node costnya > acceptedWeight atau ketika sampai di node w & cost target <= acceptedWeight.


	time complexity: O((V+E)logV), priority queue pakai binary heap.

*/
func (ch *ContractedGraph) dijkstraWitnessSearch(fromNodeIDx, targetNodeIDx int32, ignoreNodeIDx int32,
	acceptedWeight float64, maxSettledNodes int, pMax float64, contracted []bool) float64 {

	visited := make(map[int32]bool)
	
	cost := make(map[int32]float64)
	pq := NewMinHeap[int32]()
	fromNode := PriorityQueueNode[int32]{Rank: 0, Item: fromNodeIDx}
	pq.Insert(fromNode)

	cost[fromNodeIDx] = 0.0
	settledNodes := 0
	for {

		smallest, _ := pq.GetMin()
		if pq.Size() == 0 || smallest.Rank > acceptedWeight {
			return math.MaxFloat64
		}

		_, ok := cost[targetNodeIDx]
		if ok && cost[targetNodeIDx] <= acceptedWeight {
			// kita found path ke target node,  bukan yang shortest, tapi cost nya <= acceptedWeight, bisa return & gak tambahkan shortcut (u,w)
			return cost[targetNodeIDx]
		}

		currItem, _ := pq.ExtractMin()

		if contracted[currItem.Item] {
			continue
		}

		if currItem.Item == targetNodeIDx {
			// found shortest path ke target node
			return cost[currItem.Item]
		}

		if currItem.Rank > pMax {
			// rank dari current node > maximum cost path dari node u ke w , dimana u adalah semua node yang terhubung ke v & (u,v) \in E dan w adalah semua node yang terhubung ke v & (v, w) \in E, kita stop search
			out := cost[targetNodeIDx]
			if out != math.MaxFloat64 {
				return out
			}
			return math.MaxFloat64
		}

		visited[currItem.Item] = true
		for _, outIDx := range ch.ContractedFirstOutEdge[currItem.Item] {
			neighbor := ch.ContractedOutEdges[outIDx]
			if visited[neighbor.ToNodeIDX] || neighbor.ToNodeIDX == ignoreNodeIDx ||
				contracted[neighbor.ToNodeIDX] {
				continue
			}

			newCost := cost[currItem.Item] + neighbor.Weight
			neighborNode := PriorityQueueNode[int32]{Rank: newCost, Item: neighbor.ToNodeIDX}

			_, ok := cost[neighbor.ToNodeIDX]
			if !ok {
				cost[neighbor.ToNodeIDX] = newCost
				pq.Insert(neighborNode)

			} else if newCost < cost[neighbor.ToNodeIDX] {
				cost[neighbor.ToNodeIDX] = newCost

				neighborNode.Rank = newCost
				pq.DecreaseKey(neighborNode)
			}
		}

		settledNodes++
	}
}