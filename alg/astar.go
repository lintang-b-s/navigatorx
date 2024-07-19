package alg
// pindah ke a_astar.go

// import (
// 	"container/heap"
// )

// type Pather interface {
// 	PathNeighbors() []Pather
// 	PathNeighborCost(to Pather) float64
// 	PathEstimatedCost(to Pather) float64
// 	PathNeighborCostETA(to Pather) float64
// 	PathEstimatedCostETA(to Pather) float64
// 	GetStreetName() string
// }

// type astarNode struct {
// 	pather Pather
// 	rank   float64
// 	parent *astarNode
// 	index int
// }

// type nodeMap map[Pather]*astarNode

// func (nm nodeMap) get(p Pather) *astarNode {
// 	n, ok := nm[p]

// 	if !ok {
// 		n = &astarNode{pather: p}

// 		nm[p] = n
// 	}
// 	return n
// }

// // https://www.redblobgames.com/pathfinding/a-star/implementation.html#python-astar
// func AStarETA(from, to Pather) (path []Pather, eta float64, found bool, dist float64) {
// 	nm := nodeMap{}
// 	nq := &priorityQueue{}
// 	heap.Init(nq)
// 	fromNode := nm.get(from)
// 	heap.Push(nq, fromNode)

// 	costSoFar := make(map[Pather]float64)
// 	costSoFar[from] = 0.0
// 	distSoFar := make(map[Pather]float64)
// 	distSoFar[from] = 0.0

// 	cameFrom := make(map[Pather]Pather)
// 	cameFrom[from] = nil

// 	for {
// 		if nq.Len() == 0 {
// 			return
// 		}
// 		current := heap.Pop(nq).(*astarNode)

// 		if current == nm.get(to) {
// 			p := []Pather{}
// 			curr := current
// 			for curr.rank != 0 {
// 				p = append(p, curr.pather)
// 				curr = nm.get(cameFrom[curr.pather])
// 			}

// 			p = append(p, from)
// 			return p, costSoFar[current.pather], true, distSoFar[current.pather]
// 		}

// 		for _, neighbor := range current.pather.PathNeighbors() {
// 			newCost := costSoFar[current.pather] + current.pather.PathNeighborCostETA(neighbor)
// 			dist := distSoFar[current.pather] + current.pather.PathNeighborCost(neighbor)
// 			neighborNode := nm.get(neighbor)
// 			_, ok := costSoFar[neighbor]
// 			if !ok || newCost < costSoFar[neighbor] {
// 				costSoFar[neighbor] = newCost
// 				distSoFar[neighbor] = dist
// 				cameFrom[neighbor] = current.pather
// 				priority := newCost + neighbor.PathEstimatedCostETA(to)
// 				neighborNode.rank = priority
// 				heap.Push(nq, neighborNode)
// 			}
// 		}
// 	}
// }
