package alg

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

// urutin field descending by size, biar heap sizenya kecil (https://medium.com/@ali.can/memory-optimization-in-go-23a56544ccc0)
type EdgePair struct {
	Weight         float32
	Dist           float32
	ToNodeIDX      int32
	IsShortcut     bool
	RemovedEdgeOne *EdgePair
	RemovedEdgeTwo *EdgePair
}

type CHNode struct {
	OutEdges     []EdgePair
	InEdges      []EdgePair
	Lat          float32
	Lon          float32
	orderPos     int64
	IDx          int32
	StreetName   string
	TrafficLight bool
}

type Metadata struct {
	MeanDegree       float64
	ShortcutsCount   int64
	degrees          []int
	InEdgeOrigCount  []int
	OutEdgeOrigCount []int
	EdgeCount        int
	NodeCount        int
}

type ContractedGraph struct {
	Metadata   Metadata
	Ready      bool
	Rtree      *Rtree
	OrigGraph  []CHNode // graph contraction hiearchies
	AStarGraph []*CHNode

	PQNodeOrdering *priorityQueueNodeOrdering
}

var maxPollFactorHeuristic = float64(5)
var maxPollFactorContraction = float64(200) //float64(200)

func NewContractedGraph() *ContractedGraph {
	return &ContractedGraph{
		OrigGraph:  make([]CHNode, 0),
		AStarGraph: make([]*CHNode, 0),
		Ready:      false,
	}
}

func (ch *ContractedGraph) InitCHGraph(nodes []Node, edgeCount int) map[int64]int32 {
	gLen := len(nodes)
	count := 0
	var nodeIdxMap = make(map[int64]int32)

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

		ch.OrigGraph = append(ch.OrigGraph, CHNode{
			OutEdges:     []EdgePair{},
			InEdges:      []EdgePair{},
			IDx:          int32(count),
			Lat:          float32(node.Lat),
			Lon:          float32(node.Lon),
			StreetName:   node.StreetName,
			TrafficLight: node.TrafficLight,
		})

		ch.AStarGraph = append(ch.AStarGraph, &CHNode{
			OutEdges:     []EdgePair{},
			IDx:          int32(count),
			Lat:          float32(node.Lat),
			Lon:          float32(node.Lon),
			StreetName:   node.StreetName,
			TrafficLight: node.TrafficLight,
		})

		nodeIdxMap[node.ID] = int32(count)
		count++
	}
	bar.Add(1)
	ch.Metadata.degrees = make([]int, gLen)
	ch.Metadata.InEdgeOrigCount = make([]int, gLen)
	ch.Metadata.OutEdgeOrigCount = make([]int, gLen)
	ch.Metadata.ShortcutsCount = 0
	ch.PQNodeOrdering = &priorityQueueNodeOrdering{}

	// init graph original
	for idx, node := range nodes {
		outEdgeCounter := 0
		for _, edge := range node.Out_to {
			maxSpeed := edge.MaxSpeed * 1000 / 60 // m /min
			cost := edge.Cost / maxSpeed
			toIdx := nodeIdxMap[edge.To.ID]

			ch.OrigGraph[idx].OutEdges = append(ch.OrigGraph[idx].OutEdges, EdgePair{float32(cost),
				float32(edge.Cost), toIdx, false, nil, nil})
			ch.AStarGraph[idx].OutEdges = append(ch.AStarGraph[idx].OutEdges, EdgePair{float32(cost),
				float32(edge.Cost), toIdx, false, nil, nil})

			// tambah degree nodenya
			ch.Metadata.degrees[idx]++
			outEdgeCounter++
		}
		ch.Metadata.OutEdgeOrigCount[idx] = outEdgeCounter
	}

	bar.Add(1)
	// init graph edge dibalik
	for i := 0; i < gLen; i++ {
		inEdgeCounter := 0
		for _, edge := range ch.OrigGraph[i].OutEdges {
			to := edge.ToNodeIDX
			weight := edge.Weight

			ch.OrigGraph[to].InEdges = append(ch.OrigGraph[to].InEdges, EdgePair{weight,
				edge.Dist, int32(i), false, nil, nil})

			// tambah degree nodenya
			ch.Metadata.degrees[i]++
			inEdgeCounter++
		}
		ch.Metadata.InEdgeOrigCount[i] = inEdgeCounter
	}

	bar.Add(1)

	// ch.ContractedGraph = ch.OrigGraph
	ch.Metadata.EdgeCount = edgeCount
	ch.Metadata.NodeCount = gLen
	ch.Metadata.MeanDegree = float64(edgeCount) * 1.0 / float64(gLen)

	return nodeIdxMap
}

/*
referensi:
- https://github.com/graphhopper/graphhopper/blob/master/core/src/main/java/com/graphhopper/routing/ch/NodeBasedNodeContractor.java
- https://github.com/vlarmet/cppRouting/blob/master/src/contract.cpp
- https://github.com/navjindervirdee/Advanced-Shortest-Paths-Algorithms/blob/master/Contraction%20Hierarchies/DistPreprocessSmall.java
*/
func (ch *ContractedGraph) Contraction() {
	st := time.Now()
	ch.UpdatePrioritiesOfRemainingNodes() // bikin node ordering

	level := 0
	contracted := make([]bool, ch.Metadata.NodeCount)
	orderNum := int64(0)

	nq := ch.PQNodeOrdering

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
		priority := ch.calculatePriority(polledNode.NodeIDx, contracted)
		pq := *nq
		if len(pq) > 0 && priority > pq[0].rank {
			// & priority >  pq[0].rank
			// current node importantnya lebih tinggi dari next pq item
			heap.Push(nq, &pqCHNode{NodeIDx: polledNode.NodeIDx, rank: priority})
			continue
		}

		ch.OrigGraph[polledNode.NodeIDx].orderPos = orderNum

		ch.contractNode(polledNode.NodeIDx, level, contracted[polledNode.NodeIDx], contracted)
		contracted[polledNode.NodeIDx] = true
		level++
		orderNum++
		bar.Add(1)
	}
	fmt.Println("")
	fmt.Println("total shortcuts dibuat: ", ch.Metadata.ShortcutsCount)

	ch.PQNodeOrdering = nil
	ch.Metadata = Metadata{}
	runtime.GC()
	runtime.GC()
	end := time.Now().Sub(st)
	fmt.Println("lama preprocessing contraction hierarchies: : ", end.Minutes(), " menit")
}

func (ch *ContractedGraph) contractNode(nodeIDx int32, level int, isContracted bool, contracted []bool) {

	if isContracted {
		return
	}
	ch.contractNodeNow(nodeIDx, contracted)

}

func (ch *ContractedGraph) contractNodeNow(nodeIDx int32, contracted []bool) {
	degree, _, _, _ := ch.findAndHandleShortcuts(nodeIDx, ch.addOrUpdateShortcut, int(ch.Metadata.MeanDegree*maxPollFactorContraction),
		contracted)
	ch.Metadata.MeanDegree = (ch.Metadata.MeanDegree*2 + float64(degree)) / 3
	ch.disconnect(nodeIDx)

}

func (ch *ContractedGraph) findAndHandleShortcuts(nodeIDx int32, shortcutHandler func(fromNodeIDx, toNodeIDx int32, nodeIdx int32, weight float32,
	removedEdgeOne, removedEdgeTwo *EdgePair,
	outOrigEdgeCount, inOrigEdgeCount int),
	maxVisitedNodes int, contracted []bool) (int, int, int, error) {
	degree := 0
	shortcutCount := 0
	originalEdgesCount := 0
	pMax := float32(0.0)
	pInMax := float32(0.0)
	pOutMax := float32(0.0)
	for _, inEdge := range ch.OrigGraph[nodeIDx].InEdges {
		toNIDx := inEdge.ToNodeIDX
		if contracted[toNIDx] {
			continue
		}
		if inEdge.Weight > pInMax {
			pInMax = inEdge.Weight
		}
	}
	for _, outEdge := range ch.OrigGraph[nodeIDx].OutEdges {
		toNIDx := outEdge.ToNodeIDX
		if contracted[toNIDx] {
			continue
		}
		if outEdge.Weight > pOutMax {
			pOutMax = outEdge.Weight
		}
	}
	pMax = pInMax + pOutMax

	for _, inEdge := range ch.OrigGraph[nodeIDx].InEdges {
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

		for _, outEdge := range ch.OrigGraph[nodeIDx].OutEdges {
			toNode := outEdge.ToNodeIDX
			if contracted[toNode] {
				continue
			}

			if toNode == fromNodeIDx {
				// gak perlu search untuk witness dari node balik ke node itu lagi
				continue
			}

			existingDirectWeight := incomingEdgeWeight + outEdge.Weight

			maxWeight := ch.dijkstraWitnessSearch(fromNodeIDx, toNode, nodeIDx, existingDirectWeight, maxVisitedNodes, pMax,
				contracted)

			if maxWeight <= existingDirectWeight {
				// FOUND witness shortest path, tidak perlu add shortcut
				continue
			}
			// kalo d(u,w) > Pw , tambah shortcut
			// Pw = existingDirectWeight = d(u,v) + d(v,w)
			shortcutCount++
			originalEdgesCount += ch.Metadata.InEdgeOrigCount[nodeIDx] + ch.Metadata.OutEdgeOrigCount[nodeIDx]
			shortcutHandler(fromNodeIDx, toNode, nodeIDx, existingDirectWeight, &inEdge, &outEdge,
				ch.Metadata.OutEdgeOrigCount[nodeIDx], ch.Metadata.InEdgeOrigCount[nodeIDx])

		}
	}
	return degree, shortcutCount, originalEdgesCount, nil
}

func (ch *ContractedGraph) disconnect(nodeIDx int32) {
	// gak usah dihapus edge nya , biar map matching nya bener
	ch.removeContractedNode(nodeIDx)
}

func countShortcut(fromNodeIDx, toNodeIDx int32, nodeIDx int32, weight float32, removedEdgeOne, removedEdgeTwo *EdgePair,
	outOrigEdgeCount, inOrigEdgeCount int) {
	// shortcutCount++
}

func (ch *ContractedGraph) addOrUpdateShortcut(fromNodeIDx, toNodeIDx int32, nodeIDx int32, weight float32, removedEdgeOne, removedEdgeTwo *EdgePair,
	outOrigEdgeCount, inOrigEdgeCount int) {

	exists := false
	for _, edge := range ch.OrigGraph[fromNodeIDx].OutEdges {
		if edge.ToNodeIDX != toNodeIDx || !edge.IsShortcut {
			continue
		}
		exists = true
		if weight < edge.Weight {
			edge.Weight = weight
		}
	}

	for _, edge := range ch.OrigGraph[toNodeIDx].InEdges {
		if edge.ToNodeIDX != fromNodeIDx || !edge.IsShortcut {
			continue
		}
		exists = true
		if weight < edge.Weight {
			edge.Weight = weight
		}
	}

	if !exists {
		ch.addShortcut(fromNodeIDx, toNodeIDx, weight, removedEdgeOne, removedEdgeTwo)
		ch.Metadata.ShortcutsCount++
	}
}

func (ch *ContractedGraph) addShortcut(fromNodeIDx, toNodeIDx int32, weight float32, removedEdgeOne, removedEdgeTwo *EdgePair) {

	fromN := ch.OrigGraph[fromNodeIDx]
	toN := ch.OrigGraph[toNodeIDx]
	fromLoc := NewLocation(float64(fromN.Lat), float64(fromN.Lon))
	toLoc := NewLocation(float64(toN.Lat), float64(toN.Lon))
	dist := HaversineDistance(fromLoc, toLoc)
	// add shortcut outcoming edge
	dup := false
	// newETA := removedEdgeOne.ETA + removedEdgeTwo.ETA
	for _, edge := range ch.OrigGraph[fromNodeIDx].OutEdges {
		if edge.ToNodeIDX == toNodeIDx && edge.Weight > weight {
			edge.Weight = weight
			edge.Dist = float32(dist)
			// edge.ETA = newETA
			edge.RemovedEdgeOne = removedEdgeOne
			edge.RemovedEdgeTwo = removedEdgeTwo
			dup = true
			break
		}
	}
	if !dup {
		ch.OrigGraph[fromNodeIDx].OutEdges = append(ch.OrigGraph[fromNodeIDx].OutEdges, EdgePair{weight, float32(dist), toNodeIDx, true,
			removedEdgeOne, removedEdgeTwo})
		ch.Metadata.degrees[fromNodeIDx]++
	}

	dup = false
	// add shortcut incoming edge
	for _, edge := range ch.OrigGraph[toNodeIDx].InEdges {
		if edge.ToNodeIDX == fromNodeIDx && edge.Weight > weight {
			edge.Weight = weight
			edge.Dist = float32(dist)
			// edge.ETA = newETA
			edge.RemovedEdgeOne = removedEdgeTwo
			edge.RemovedEdgeTwo = removedEdgeOne
			dup = true
			break
		}

	}
	if !dup {
		ch.OrigGraph[toNodeIDx].InEdges = append(ch.OrigGraph[toNodeIDx].InEdges, EdgePair{weight, float32(dist), fromNodeIDx, true, removedEdgeOne,
			removedEdgeTwo,
		})
		ch.Metadata.degrees[toNodeIDx]++

	}
}

func (ch *ContractedGraph) removeContractedNode(nodeIDx int32) {

	// remove semua incoming edge ke nodeIDx
	for _, nEdge := range ch.OrigGraph[nodeIDx].InEdges {
		nd := nEdge.ToNodeIDX
		ind := []int{}
		for i, inEdge := range ch.OrigGraph[nd].OutEdges {
			// incoming edge ke nodeIDx
			if inEdge.ToNodeIDX == nodeIDx && !inEdge.IsShortcut {
				ind = append(ind, i)

			}

		}
		ind = reverse(ind)
		for _, edgeIDx := range ind {
			quickDelete(edgeIDx, &ch.OrigGraph[nd], "f")
			ch.Metadata.degrees[nd]--
			ch.Metadata.OutEdgeOrigCount[nd]-- // outgoing edge dari nd berkurang 1
		}
	}

	// remove semua outgoing edge dari nodeIDx
	for _, nEdge := range ch.OrigGraph[nodeIDx].OutEdges {
		nd := nEdge.ToNodeIDX
		ind := []int{}
		for i, outEdge := range ch.OrigGraph[nd].InEdges {
			// outgoing edge dari nodeIDx
			if outEdge.ToNodeIDX == nodeIDx && !outEdge.IsShortcut {
				ind = append(ind, i)
			}

		}
		ind = reverse(ind)
		for _, edgeIDx := range ind {
			quickDelete(edgeIDx, &ch.OrigGraph[nd], "b")
			ch.Metadata.degrees[nd]--
			ch.Metadata.InEdgeOrigCount[nd]-- // incoming edge ke nd berkurang 1
		}

	}

	ch.Metadata.degrees[nodeIDx] = 0
	ch.Metadata.InEdgeOrigCount[nodeIDx] = 0
	ch.Metadata.OutEdgeOrigCount[nodeIDx] = 0
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

func (ch *ContractedGraph) calculatePriority(nodeIDx int32, contracted []bool) float64 {

	_, shortcutsCount, originalEdgesCount, _ := ch.findAndHandleShortcuts(nodeIDx, countShortcut, int(ch.Metadata.MeanDegree*maxPollFactorHeuristic),
		contracted)

	// |shortcuts(v)| − |{(u, v) | v uncontracted}| − |{(v, w) | v uncontracted}|
	// outDegree+inDegree
	edgeDifference := float64(shortcutsCount - ch.Metadata.degrees[nodeIDx])
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

func (ch *ContractedGraph) UpdatePrioritiesOfRemainingNodes() {
	heap.Init(ch.PQNodeOrdering)
	bar := progressbar.NewOptions(len(ch.OrigGraph),
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

	contracted := make([]bool, ch.Metadata.NodeCount)
	for nodeIDx, _ := range ch.OrigGraph {
		// if (isContracted(node)) {
		// 	continue
		// }
		priority := ch.calculatePriority(int32(nodeIDx), contracted)
		heap.Push(ch.PQNodeOrdering, &pqCHNode{NodeIDx: int32(nodeIDx), rank: float64(priority)})
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
	maxSpeed := 90.0 * 1000.0 / 60.0                               // m/min
	r := math.Sqrt(float64(absLatSq+absLonSq)) * 100000 / maxSpeed // * 100000 -> meter
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

func (ch *ContractedGraph) IsChReady() bool {
	return ch.Ready
}
