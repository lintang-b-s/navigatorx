package alg

import (
	"container/heap"
	"errors"
	"fmt"
	"math"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

type EdgePair struct {
	ToNodeIDX      int32
	Weight         float64
	Dist           float64
	ETA            float64
	IsShortcut     bool
	RemovedEdgeOne *EdgePair
	RemovedEdgeTwo *EdgePair
}

type CHNode struct {
	IDx          int32
	Lat          float64
	Lon          float64
	StreetName   string
	TrafficLight bool
	orderPos     int64
	OutEdges     []EdgePair
	InEdges      []EdgePair
}

type Metadata struct {
	MeanDegree       float64
	EdgeCount        int
	NodeCount        int
	degrees          []int
	InEdgeOrigCount  []int
	OutEdgeOrigCount []int
	ShortcutsCount   int64
}

type ContractedGraph struct {
	OrigGraph  []*CHNode // graph contraction hiearchies
	AStarGraph []*CHNode

	// RevGraph       []*CHNode // graph dengan reversed edge
	Metadata       Metadata
	PQNodeOrdering *priorityQueueNodeOrdering
}

var maxPollFactorHeuristic = float64(5)
var maxPollFactorContraction = float64(200) //float64(200)
// var CHGraph = ContractedGraph{OrigGraph: []*CHNode{}, RevGraph: []*CHNode{}}
var CHGraph = ContractedGraph{OrigGraph: []*CHNode{}, AStarGraph: []*CHNode{}}

// var CHGraph = ContractedGraph{OrigGraph: []*CHNode{}, }

var NodeIdxMap = make(map[int64]int32)

func InitCHGraph(nodes []Node, edgeCount int) {
	gLen := len(nodes)
	count := 0

	bar := progressbar.NewOptions(3,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][3/6][reset] Membuat graph..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	for _, node := range nodes {

		CHGraph.OrigGraph = append(CHGraph.OrigGraph, &CHNode{
			OutEdges:     []EdgePair{},
			InEdges:      []EdgePair{},
			IDx:          int32(count),
			Lat:          node.Lat,
			Lon:          node.Lon,
			StreetName:   node.StreetName,
			TrafficLight: node.TrafficLight,
		})

		// CHGraph.AStarGraph = append(CHGraph.AStarGraph, &CHNode{
		// 	OutEdges:     []EdgePair{},
		// 	IDx:          int32(count),
		// 	Lat:          node.Lat,
		// 	Lon:          node.Lon,
		// 	StreetName:   node.StreetName,
		// 	TrafficLight: node.TrafficLight,
		// })

		// CHGraph.RevGraph = append(CHGraph.RevGraph, &CHNode{Edges: []EdgePair{},
		// 	IDx:          int32(count),
		// 	Lat:          node.Lat,
		// 	Lon:          node.Lon,
		// 	StreetName:   node.StreetName,
		// 	TrafficLight: node.TrafficLight,
		// })
		NodeIdxMap[node.ID] = int32(count)
		count++
	}
	bar.Add(1)
	CHGraph.Metadata.degrees = make([]int, gLen)
	CHGraph.Metadata.InEdgeOrigCount = make([]int, gLen)
	CHGraph.Metadata.OutEdgeOrigCount = make([]int, gLen)
	CHGraph.Metadata.ShortcutsCount = 0
	CHGraph.PQNodeOrdering = &priorityQueueNodeOrdering{}

	// init graph original
	for idx, node := range nodes {
		outEdgeCounter := 0
		for _, edge := range node.Out_to {
			maxSpeed := edge.MaxSpeed * 1000 / 60 // m /min
			cost := edge.Cost / maxSpeed
			toIdx := NodeIdxMap[edge.To.ID]

			CHGraph.OrigGraph[idx].OutEdges = append(CHGraph.OrigGraph[idx].OutEdges, EdgePair{toIdx, cost,
				edge.Cost, cost, false, nil, nil})
			// CHGraph.AStarGraph[idx].OutEdges = append(CHGraph.AStarGraph[idx].OutEdges, EdgePair{toIdx, cost,
			// 	edge.Cost, cost, false, nil, nil})

			// tambah degree nodenya
			CHGraph.Metadata.degrees[idx]++
			outEdgeCounter++
		}
		CHGraph.Metadata.OutEdgeOrigCount[idx] = outEdgeCounter
	}

	bar.Add(1)
	// init graph edge dibalik
	for i := 0; i < gLen; i++ {
		inEdgeCounter := 0
		for _, edge := range CHGraph.OrigGraph[i].OutEdges {
			to := edge.ToNodeIDX
			weight := edge.Weight

			CHGraph.OrigGraph[to].InEdges = append(CHGraph.OrigGraph[to].InEdges, EdgePair{int32(i), weight,
				edge.Dist, weight, false, nil, nil})

			// tambah degree nodenya
			CHGraph.Metadata.degrees[i]++
			inEdgeCounter++
		}
		CHGraph.Metadata.InEdgeOrigCount[i] = inEdgeCounter
	}

	bar.Add(1)

	// CHGraph.ContractedGraph = CHGraph.OrigGraph
	CHGraph.Metadata.EdgeCount = edgeCount
	CHGraph.Metadata.NodeCount = gLen
	CHGraph.Metadata.MeanDegree = float64(edgeCount) * 1.0 / float64(gLen)
}

/*
referensi: 
- https://github.com/graphhopper/graphhopper/blob/master/core/src/main/java/com/graphhopper/routing/ch/NodeBasedNodeContractor.java
- https://github.com/vlarmet/cppRouting/blob/master/src/contract.cpp
- https://github.com/navjindervirdee/Advanced-Shortest-Paths-Algorithms/blob/master/Contraction%20Hierarchies/DistPreprocessSmall.java

*/
func Contraction() {

	UpdatePrioritiesOfRemainingNodes() // bikin node ordering

	level := 0
	contracted := make([]bool, CHGraph.Metadata.NodeCount)
	orderNum := int64(0)

	nq := CHGraph.PQNodeOrdering

	bar := progressbar.NewOptions(nq.Len(),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][6/6][reset] Membuat contracted graph (contraction hiearchies)..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	for nq.Len() != 0 {
		polledNode := heap.Pop(nq).(*pqCHNode)

		// lazy update
		priority := calculatePriority(polledNode.NodeIDx, contracted)
		pq := *nq
		if len(pq) > 0 && priority > pq[0].rank {
			// & priority >  pq[0].rank
			// current node importantnya lebih tinggi dari next pq item
			heap.Push(nq, &pqCHNode{NodeIDx: polledNode.NodeIDx, rank: priority})
			continue
		}

		CHGraph.OrigGraph[polledNode.NodeIDx].orderPos = orderNum

		contractNode(polledNode.NodeIDx, level, contracted[polledNode.NodeIDx], contracted)
		contracted[polledNode.NodeIDx] = true
		level++
		orderNum++
		bar.Add(1)
	}
	fmt.Println("")
	fmt.Println("total shortcuts dibuat: ", CHGraph.Metadata.ShortcutsCount)

	CHGraph.PQNodeOrdering = nil
}

func contractNode(nodeIDx int32, level int, isContracted bool, contracted []bool) {

	if isContracted {
		return
	}
	contractNodeNow(nodeIDx, contracted)

}

func contractNodeNow(nodeIDx int32, contracted []bool) {
	degree, _, _, _ := findAndHandleShortcuts(nodeIDx, addOrUpdateShortcut, int(CHGraph.Metadata.MeanDegree*maxPollFactorContraction),
		contracted)
	CHGraph.Metadata.MeanDegree = (CHGraph.Metadata.MeanDegree*2 + float64(degree)) / 3 // (CHGraph.Metadata.MeanDegree*2 + float64(degree)) / 3
	// if shortcutCount > 0 {
	disconnect(nodeIDx)
	// }
}

func findAndHandleShortcuts(nodeIDx int32, shortcutHandler func(fromNodeIDx, toNodeIDx int32, weight float64,
	removedEdgeOne, removedEdgeTwo *EdgePair,
	outOrigEdgeCount, inOrigEdgeCount int),
	maxVisitedNodes int, contracted []bool) (int, int, int, error) {
	degree := 0
	shortcutCount := 0
	originalEdgesCount := 0
	pMax := 0.0
	pInMax := 0.0
	pOutMax := 0.0
	for _, inEdge := range CHGraph.OrigGraph[nodeIDx].InEdges {
		toNIDx := inEdge.ToNodeIDX
		if contracted[toNIDx] {
			continue
		}
		if inEdge.Weight > pInMax {
			pInMax = inEdge.Weight
		}
	}
	for _, outEdge := range CHGraph.OrigGraph[nodeIDx].OutEdges {
		toNIDx := outEdge.ToNodeIDX
		if contracted[toNIDx] {
			continue
		}
		if outEdge.Weight > pOutMax {
			pOutMax = outEdge.Weight
		}
	}
	pMax = pInMax + pOutMax

	for _, inEdge := range CHGraph.OrigGraph[nodeIDx].InEdges {
		fromNodeIDx := inEdge.ToNodeIDX
		if fromNodeIDx == int32(nodeIDx) {
			return 0, 0, 0, errors.New(fmt.Sprintf(`unexpected loop-edge at node: %v `, nodeIDx))
		}
		if contracted[fromNodeIDx] {
			continue
		}

		incomingEdgeWeight := inEdge.Weight

		// outging edge dari nodeIDx
		degree++

		for _, outEdge := range CHGraph.OrigGraph[nodeIDx].OutEdges {
			toNode := outEdge.ToNodeIDX
			if contracted[toNode] {
				continue
			}

			if toNode == fromNodeIDx {
				// gak perlu search untuk witness dari node balik ke node itu lagi
				continue
			}

			existingDirectWeight := incomingEdgeWeight + outEdge.Weight

			maxWeight := dijkstraWitnessSearch(fromNodeIDx, toNode, nodeIDx, existingDirectWeight, maxVisitedNodes, pMax,
				contracted)

			if maxWeight <= existingDirectWeight {
				// FOUND witness path, tidak perlu add shortcut
				continue
			}
			// kalo d(u,w) > Pw , tambah shortcut
			// Pw = existingDirectWeight = d(u,v) + d(v,w)
			shortcutCount++
			originalEdgesCount += CHGraph.Metadata.InEdgeOrigCount[nodeIDx] + CHGraph.Metadata.OutEdgeOrigCount[nodeIDx]
			shortcutHandler(fromNodeIDx, toNode, existingDirectWeight, &inEdge, &outEdge,
				CHGraph.Metadata.OutEdgeOrigCount[nodeIDx], CHGraph.Metadata.InEdgeOrigCount[nodeIDx])

		}
	}
	return degree, shortcutCount, originalEdgesCount, nil
}

func disconnect(nodeIDx int32) {
	// gak usah dihapus edge nya , biar map matching nya bener
	removeContractedNode(nodeIDx)
}

func countShortcut(fromNodeIDx, toNodeIDx int32, weight float64, removedEdgeOne, removedEdgeTwo *EdgePair,
	outOrigEdgeCount, inOrigEdgeCount int) {
	// shortcutCount++
}

func addOrUpdateShortcut(fromNodeIDx, toNodeIDx int32, weight float64, removedEdgeOne, removedEdgeTwo *EdgePair,
	outOrigEdgeCount, inOrigEdgeCount int) {
	exists := false
	for _, edge := range CHGraph.OrigGraph[fromNodeIDx].OutEdges {
		if edge.ToNodeIDX != toNodeIDx || !edge.IsShortcut {
			continue
		}
		exists = true
		if weight < edge.Weight {
			edge.Weight = weight
		}
	}

	for _, edge := range CHGraph.OrigGraph[toNodeIDx].InEdges {
		if edge.ToNodeIDX != fromNodeIDx || !edge.IsShortcut {
			continue
		}
		exists = true
		if weight < edge.Weight {
			edge.Weight = weight
		}
	}

	if !exists {
		addShortcut(fromNodeIDx, toNodeIDx, weight, removedEdgeOne, removedEdgeTwo)
		CHGraph.Metadata.ShortcutsCount++
	}
}

func addShortcut(fromNodeIDx, toNodeIDx int32, weight float64, removedEdgeOne, removedEdgeTwo *EdgePair) {
	fromN := CHGraph.OrigGraph[fromNodeIDx]
	toN := CHGraph.OrigGraph[toNodeIDx]
	fromLoc := NewLocation(fromN.Lat, fromN.Lon)
	toLoc := NewLocation(toN.Lat, toN.Lon)
	dist := HaversineDistance(fromLoc, toLoc)
	// add shortcut outcoming edge
	dup := false
	// newETA := removedEdgeOne.ETA + removedEdgeTwo.ETA
	for _, edge := range CHGraph.OrigGraph[fromNodeIDx].OutEdges {
		if edge.ToNodeIDX == toNodeIDx && edge.Weight > weight {
			edge.Weight = weight
			edge.Dist = dist
			// edge.ETA = newETA
			edge.RemovedEdgeOne = removedEdgeOne
			edge.RemovedEdgeTwo = removedEdgeTwo
			dup = true
			break
		}
	}
	if !dup {
		CHGraph.OrigGraph[fromNodeIDx].OutEdges = append(CHGraph.OrigGraph[fromNodeIDx].OutEdges, EdgePair{toNodeIDx, weight, dist, weight, true,
			removedEdgeOne, removedEdgeTwo})
	}

	dup = false
	// add shortcut incoming edge
	for _, edge := range CHGraph.OrigGraph[toNodeIDx].InEdges {
		if edge.ToNodeIDX == fromNodeIDx && edge.Weight > weight {
			edge.Weight = weight
			edge.Dist = dist
			// edge.ETA = newETA
			edge.RemovedEdgeOne = removedEdgeTwo
			edge.RemovedEdgeTwo = removedEdgeOne
			dup = true
			break
		}

	}
	if !dup {
		CHGraph.OrigGraph[toNodeIDx].InEdges = append(CHGraph.OrigGraph[toNodeIDx].InEdges, EdgePair{fromNodeIDx, weight, dist, weight, true, removedEdgeOne,
			removedEdgeTwo,
		})
	}
}

func removeContractedNode(nodeIDx int32) {

	// remove semua incoming edge ke nodeIDx
	for _, nEdge := range CHGraph.OrigGraph[nodeIDx].InEdges {
		nd := nEdge.ToNodeIDX
		ind := []int{}
		for i, inEdge := range CHGraph.OrigGraph[nd].OutEdges {
			//incoming edge ke nodeIDx
			if inEdge.ToNodeIDX == nodeIDx && !inEdge.IsShortcut {
				ind = append(ind, i)

			}
		}
		ind = reverse(ind)
		for _, edgeIDx := range ind {
			quickDelete(edgeIDx, CHGraph.OrigGraph[nd], "f")
			CHGraph.Metadata.degrees[nd]--
			CHGraph.Metadata.OutEdgeOrigCount[nd]-- // outgoing edge dari nd berkurang 1
		}
	}

	// remove semua outgoing edge dari nodeIDx
	for _, nEdge := range CHGraph.OrigGraph[nodeIDx].OutEdges {
		nd := nEdge.ToNodeIDX
		ind := []int{}
		for i, outEdge := range CHGraph.OrigGraph[nd].InEdges {
			//outgoing edge dari nodeIDx
			if outEdge.ToNodeIDX == nodeIDx && !outEdge.IsShortcut {
				ind = append(ind, i)
			}
		}
		ind = reverse(ind)
		for _, edgeIDx := range ind {
			quickDelete(edgeIDx, CHGraph.OrigGraph[nd], "b")
			CHGraph.Metadata.degrees[nd]--
			CHGraph.Metadata.InEdgeOrigCount[nd]-- // incoming edge ke nd berkurang 1
		}
	}

	//  edges dari nodeIdx bukan nil
	// nodeIdx bukan nil
	CHGraph.Metadata.degrees[nodeIDx] = 0
	CHGraph.Metadata.InEdgeOrigCount[nodeIDx] = 0
	CHGraph.Metadata.OutEdgeOrigCount[nodeIDx] = 0
}

func quickDelete(idx int, g *CHNode, dir string) {
	if dir == "f" {
		g.OutEdges[idx] = g.OutEdges[len(g.OutEdges)-1]
		g.OutEdges = g.OutEdges[:len(g.OutEdges)-1]
	} else {
		g.InEdges[idx] = g.InEdges[len(g.InEdges)-1]
		g.InEdges = g.InEdges[:len(g.InEdges)-1]
	}

}

func calculatePriority(nodeIDx int32, contracted []bool) float64 {

	_, shortcutsCount, originalEdgesCount, _ := findAndHandleShortcuts(nodeIDx, countShortcut, int(CHGraph.Metadata.MeanDegree*maxPollFactorHeuristic),
		contracted)

	// |shortcuts(v)| − |{(u, v) | v uncontracted}| − |{(v, w) | v uncontracted}|
	// outDegree+inDegree
	edgeDifference := float64(shortcutsCount - CHGraph.Metadata.degrees[nodeIDx])
	// 1 shortcut, 10 degree = -9
	// 5 shortcut, 10 degree = -5
	// 5 shortcut ED > 1 shortcut ED

	return 10*edgeDifference + 1*float64(originalEdgesCount)
}

type pqCHNode struct {
	NodeIDx int32
	rank    float64
	index   int
}

func UpdatePrioritiesOfRemainingNodes() {
	heap.Init(CHGraph.PQNodeOrdering)
	bar := progressbar.NewOptions(len(CHGraph.OrigGraph),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()), //you should install "github.com/k0kubun/go-ansi"
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][5/6][reset] Membuat node ordering (contraction hiearchies)..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	contracted := make([]bool, CHGraph.Metadata.NodeCount)
	for nodeIDx, _ := range CHGraph.OrigGraph {
		// if (isContracted(node)) {
		// 	continue
		// }
		priority := calculatePriority(int32(nodeIDx), contracted)
		heap.Push(CHGraph.PQNodeOrdering, &pqCHNode{NodeIDx: int32(nodeIDx), rank: float64(priority)})
		bar.Add(1)
	}
	fmt.Println("")
}

func (n *CHNode) PathEstimatedCostETA(to CHNode) float64 {

	toN := to
	absLat := toN.Lat - n.Lat
	if absLat < 0 {
		absLat = -absLat
	}
	absLon := toN.Lon - n.Lon
	if absLon < 0 {
		absLon = -absLon
	}

	absLatSq := absLat * absLat
	absLonSq := absLon * absLon

	// r := float64(absLat + absLon)
	maxSpeed := 90.0 * 1000.0 / 60.0                      // m/min
	r := math.Sqrt(absLatSq+absLonSq) * 100000 / maxSpeed // * 100000 -> meter
	return r
}

func reverseCH(p []CHNode) []CHNode {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
	return p
}

func reverse(arr []int) []int {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// func findAndHandleShortcuts(nodeIDx int32, shortcutHandler func(fromNodeIDx, toNodeIDx int32, weight float64,
// 	outOrigEdgeCount, inOrigEdgeCount int), maxVisitedNodes int) (int, int, int, error) {
// 	degree := 0
// 	shortcutCount := 0
// 	originalEdgesCount := 0

// 	for _, incEdge := range CHGraph.RevGraph[nodeIDx].Edges {
// 		fromNodeIDx := incEdge.ToNodeIDX
// 		if fromNodeIDx == int32(nodeIDx) {
// 			return 0, 0, 0, errors.New(fmt.Sprintf(`unexpected loop-edge at node: %v `, nodeIDx))
// 		}

// 		incomingEdgeWeight := incEdge.Weight

// 		// outging edge dari nodeIDx
// 		degree++

// 		for _, outEdge := range CHGraph.OrigGraph[nodeIDx].Edges {
// 			toNode := outEdge.ToNodeIDX

// 			if toNode == fromNodeIDx {
// 				// gak perlu search untuk witness dari node balik ke node itu lagi
// 				continue
// 			}

// 			existingDirectWeight := incomingEdgeWeight + outEdge.Weight

// 			maxWeight := dijkstraWitnessSearch(fromNodeIDx, toNode, nodeIDx, existingDirectWeight, maxVisitedNodes)

// 			if maxWeight <= existingDirectWeight {
// 				// FOUND witness path, tidak perlu add shortcut
// 				continue
// 			}
// 			shortcutCount++
// 			originalEdgesCount += CHGraph.Metadata.InEdgeOrigCount[nodeIDx] + CHGraph.Metadata.OutEdgeOrigCount[nodeIDx]
// 			shortcutHandler(fromNodeIDx, toNode, existingDirectWeight,
// 				CHGraph.Metadata.OutEdgeOrigCount[nodeIDx], CHGraph.Metadata.InEdgeOrigCount[nodeIDx])

// 		}
// 	}
// 	return degree, shortcutCount, originalEdgesCount, nil
// }

// func addOrUpdateShortcut(fromNodeIDx, toNodeIDx int32, weight float64,
// 	outOrigEdgeCount, inOrigEdgeCount int) {
// 	exists := false
// 	for _, edge := range CHGraph.OrigGraph[fromNodeIDx].Edges {
// 		if edge.ToNodeIDX != toNodeIDx || !edge.IsShortcut {
// 			continue
// 		}
// 		exists = true
// 		if weight < edge.Weight {
// 			edge.Weight = weight
// 		}
// 	}

// 	for _, edge := range CHGraph.RevGraph[toNodeIDx].Edges {
// 		if edge.ToNodeIDX != fromNodeIDx || !edge.IsShortcut {
// 			continue
// 		}
// 		exists = true
// 		if weight < edge.Weight {
// 			edge.Weight = weight
// 		}
// 	}

// 	if !exists {
// 		addShortcut(fromNodeIDx, toNodeIDx, weight)
// 	}
// }

// func addShortcut(fromNodeIDx, toNodeIDx int32, weight float64) {
// 	fromN := CHGraph.OrigGraph[fromNodeIDx]
// 	toN := CHGraph.OrigGraph[toNodeIDx]
// 	fromLoc := NewLocation(fromN.Lat, fromN.Lon)
// 	toLoc := NewLocation(toN.Lat, toN.Lon)
// 	dist := HaversineDistance(fromLoc, toLoc)
// 	// add shortcut outcoming edge
// 	dup := false
// 	for _, edge := range CHGraph.OrigGraph[fromNodeIDx].Edges {
// 		if edge.ToNodeIDX == toNodeIDx && edge.Weight > weight {
// 			edge.Weight = weight
// 			edge.Dist = dist
// 			dup = true
// 			break
// 		}
// 	}
// 	if !dup {
// 		CHGraph.OrigGraph[fromNodeIDx].Edges = append(CHGraph.OrigGraph[fromNodeIDx].Edges, EdgePair{toNodeIDx, weight, dist, true})
// 	}

// 	dup = false
// 	// add shortcut incoming edge
// 	for _, edge := range CHGraph.RevGraph[toNodeIDx].Edges {
// 		if edge.ToNodeIDX == fromNodeIDx && edge.Weight > weight {
// 			edge.Weight = weight
// 			edge.Dist = dist
// 			dup = true
// 			break
// 		}

// 	}
// 	if !dup {
// 		CHGraph.RevGraph[toNodeIDx].Edges = append(CHGraph.RevGraph[toNodeIDx].Edges, EdgePair{fromNodeIDx, weight, dist, true})
// 	}
// }

// func removeContractedNode(nodeIDx int32) {

// 	// remove semua incoming edge ke nodeIDx
// 	for _, nEdge := range CHGraph.RevGraph[nodeIDx].Edges {
// 		nd := nEdge.ToNodeIDX
// 		ind := []int{}
// 		for i, inEdge := range CHGraph.OrigGraph[nd].Edges {
// 			//incoming edge ke nodeIDx
// 			if inEdge.ToNodeIDX == nodeIDx {
// 				ind = append(ind, i)

// 			}
// 		}
// 		ind = reverse(ind)
// 		for _, edgeIDx := range ind {
// 			quickDelete(edgeIDx, CHGraph.OrigGraph[nd])
// 			CHGraph.Metadata.degrees[nd]--
// 			CHGraph.Metadata.OutEdgeOrigCount[nd]-- // outgoing edge dari nd berkurang 1
// 		}
// 	}

// 	// remove semua outgoing edge dari nodeIDx
// 	for _, nEdge := range CHGraph.OrigGraph[nodeIDx].Edges {
// 		nd := nEdge.ToNodeIDX
// 		ind := []int{}
// 		for i, outEdge := range CHGraph.RevGraph[nd].Edges {
// 			//outgoing edge dari nodeIDx
// 			if outEdge.ToNodeIDX == nodeIDx {
// 				ind = append(ind, i)
// 			}
// 		}
// 		ind = reverse(ind)
// 		for _, edgeIDx := range ind {
// 			quickDelete(edgeIDx, CHGraph.RevGraph[nd])
// 			CHGraph.Metadata.degrees[nd]--
// 			CHGraph.Metadata.InEdgeOrigCount[nd]-- // incoming edge ke nd berkurang 1
// 		}
// 	}

// 	//  edges dari nodeIdx bukan nil
// 	// nodeIdx bukan nil
// 	CHGraph.Metadata.degrees[nodeIDx] = 0
// 	CHGraph.Metadata.InEdgeOrigCount[nodeIDx] = 0
// 	CHGraph.Metadata.OutEdgeOrigCount[nodeIDx] = 0
// }

// func quickDelete(idx int, g *CHNode) {
// 	g.Edges[idx] = g.Edges[len(g.Edges)-1]
// 	g.Edges = g.Edges[:len(g.Edges)-1]
// }
