package alg

import (
	"sort"

	"github.com/dhconnelly/rtreego"
)

type NodePoint struct {
	Node CHNode
	Idx  int32
	Dist float64
}

func SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNodeIdx int32, err error) {
	wantToSnap := rtreego.Point{lat, lon}
	stNeighbors := StRTree.NearestNeighbors(4, wantToSnap)

	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

	snappedStNode := int32(0)
	best := 100000000.0

	// snap point ke  node jalan terdekat/posisi location seharusnya
	for _, st := range stNeighbors {

		street := st.(*StreetRect).Street
		nearest := CHGraph.OrigGraph[NodeIdxMap[street.NodesID[0]]]
		nearestStPoint := *nearest       // node di jalan yg paling dekat dg gps
		secondNearestStPoint := *nearest // node di jalan yang paling dekat kedua dg gps

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, nodeID := range street.NodesID {
			nodeIdx := NodeIdxMap[nodeID]
			node := CHGraph.OrigGraph[nodeIdx]
			nodeLoc := NewLocation(node.Lat, node.Lon)
			streetNodes = append(streetNodes, NodePoint{*node, nodeIdx, HaversineDistance(wantToSnapLoc, nodeLoc)})
		}

		sort.Slice(streetNodes, func(i, j int) bool {
			return streetNodes[i].Dist < streetNodes[j].Dist
		})

		nearestStPoint = streetNodes[0].Node
		nearestStNodeIdx := streetNodes[0].Idx
		secondNearestStPoint = streetNodes[1].Node

		// project point ke line segment jalan antara 2 point tadi
		projection := ProjectPointToLine(nearestStPoint, secondNearestStPoint, wantToSnap)

		projectionLoc := NewLocation(projection.Lat, projection.Lon)

		// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
		if HaversineDistance(wantToSnapLoc, projectionLoc) < best {
			best = HaversineDistance(wantToSnapLoc, projectionLoc)
			snappedStNode = nearestStNodeIdx
		}
	}

	return snappedStNode, nil
}

func SnapLocationToRoadNetworkNodeRtreeCH(lat, lon float64, dir string) (snappedRoadNodeIdx int32, err error) {
	wantToSnap := rtreego.Point{lat, lon}
	stNeighbors := StRTree.NearestNeighbors(4, wantToSnap)

	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

	snappedStNode := int32(0)
	best := 100000000.0

	// snap point ke  node jalan terdekat/posisi location seharusnya
	for _, st := range stNeighbors {

		street := st.(*StreetRect).Street
		nearest := CHGraph.OrigGraph[NodeIdxMap[street.NodesID[0]]]
		nearestStPoint := *nearest       // node di jalan yg paling dekat dg gps
		secondNearestStPoint := *nearest // node di jalan yang paling dekat kedua dg gps

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, nodeID := range street.NodesID {
			nodeIdx := NodeIdxMap[nodeID]
			node := CHGraph.OrigGraph[nodeIdx]
			nodeLoc := NewLocation(node.Lat, node.Lon)
			streetNodes = append(streetNodes, NodePoint{*node, nodeIdx, HaversineDistance(wantToSnapLoc, nodeLoc)})
		}

		sort.Slice(streetNodes, func(i, j int) bool {
			return streetNodes[i].Dist < streetNodes[j].Dist
		})

		nearestStPoint = streetNodes[0].Node
		nearestStNodeIdx := streetNodes[0].Idx
		secondNearestStPoint = streetNodes[1].Node

		// project point ke line segment jalan antara 2 point tadi
		projection := ProjectPointToLine(nearestStPoint, secondNearestStPoint, wantToSnap)

		projectionLoc := NewLocation(projection.Lat, projection.Lon)

		// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
		if dir == "f" {
			if HaversineDistance(wantToSnapLoc, projectionLoc) < best &&
				len(CHGraph.OrigGraph[nearestStNodeIdx].OutEdges) != 0 {
				best = HaversineDistance(wantToSnapLoc, projectionLoc)
				snappedStNode = nearestStNodeIdx
			}
		} else {
			if HaversineDistance(wantToSnapLoc, projectionLoc) < best &&
				len(CHGraph.OrigGraph[nearestStNodeIdx].InEdges) != 0 {
				best = HaversineDistance(wantToSnapLoc, projectionLoc)
				snappedStNode = nearestStNodeIdx
			}
		}

	}

	return snappedStNode, nil
}

// func SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNode *Node, err error) {
// 	wantToSnap := rtreego.Point{lat, lon}
// 	stNeighbors := StRTree.NearestNeighbors(3, wantToSnap)

// 	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

// 	snappedStNode := &Node{}
// 	best := 100000000.0

// 	// snap point ke  node jalan terdekat/posisi location seharusnya
// 	for _, st := range stNeighbors {

// 		street := st.(*StreetRect).Street
// 		nearest := NodeIdxMap[street.NodesID[0] ]
// 		nearestStPoint := street.NodesID[0]       // node di jalan yg paling dekat dg gps
// 		secondNearestStPoint := street.NodesID[0] // node di jalan yang paling dekat kedua dg gps

// 		// mencari 2 point dijalan yg paling dekat dg gps
// 		streetNodes := []NodePoint{}
// 		for _, node := range street.NodesID {
// 			nodeLoc := NewLocation(node.Lat, node.Lon)
// 			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc)})
// 		}

// 		sort.Slice(streetNodes, func(i, j int) bool {
// 			return streetNodes[i].Dist < streetNodes[j].Dist
// 		})

// 		nearestStPoint = streetNodes[0].Node
// 		secondNearestStPoint = streetNodes[1].Node

// 		// project point ke line segment jalan antara 2 point tadi
// 		projection := ProjectPointToLine(*nearestStPoint, *secondNearestStPoint, wantToSnap)

// 		projectionLoc := NewLocation(projection.Lat, projection.Lon)

// 		// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
// 		if HaversineDistance(wantToSnapLoc, projectionLoc) < best {
// 			best = HaversineDistance(wantToSnapLoc, projectionLoc)
// 			snappedStNode = nearestStPoint
// 		}
// 	}

// 	return snappedStNode, nil
// }
