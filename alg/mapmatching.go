package alg

import (
	"sort"

	"github.com/dhconnelly/rtreego"
)

// urutin field struct descending by size , biar makin kecil heap size nya
type NodePoint struct {
	Node CHNode2
	Dist float64
	Idx  int32
}

func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNodeIdx int32, err error) {
	wantToSnap := rtreego.Point{lat, lon}
	stNeighbors := ch.Rtree.StRtree.NearestNeighbors(4, wantToSnap)

	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

	snappedStNode := int32(0)
	best := 100000000.0

	// snap point ke  node jalan terdekat/posisi location seharusnya
	for _, st := range stNeighbors {

		street := st.(*StreetRect).Street
		nearest := ch.ContractedNodes[street.NodesID[0]]
		nearestStPoint := nearest       // node di jalan yg paling dekat dg gps
		secondNearestStPoint := nearest // node di jalan yang paling dekat kedua dg gps

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, nodeID := range street.NodesID {
			nodeIdx := nodeID
			node := ch.ContractedNodes[nodeIdx]
			nodeLoc := NewLocation(float64(node.Lat), float64(node.Lon))
			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc), int32(nodeIdx)})
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

// func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeRtreeCH(lat, lon float64, dir string) (snappedRoadNodeIdx int32, err error) {
// 	wantToSnap := rtreego.Point{lat, lon}
// 	stNeighbors := ch.Rtree.StRtree.NearestNeighbors(4, wantToSnap)

// 	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

// 	snappedStNode := int32(0)
// 	best := 100000000.0

// 	// snap point ke  node jalan terdekat/posisi location seharusnya
// 	for _, st := range stNeighbors {

// 		street := st.(*StreetRect).Street
// 		nearest := ch.OrigGraph[street.NodesID[0]]
// 		nearestStPoint := nearest       // node di jalan yg paling dekat dg gps
// 		secondNearestStPoint := nearest // node di jalan yang paling dekat kedua dg gps

// 		// mencari 2 point dijalan yg paling dekat dg gps
// 		streetNodes := []NodePoint{}
// 		for _, nodeID := range street.NodesID {
// 			nodeIdx := nodeID
// 			node := ch.OrigGraph[nodeIdx]
// 			nodeLoc := NewLocation(float64(node.Lat), float64(node.Lon))
// 			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc), int32(nodeIdx)})
// 		}

// 		sort.Slice(streetNodes, func(i, j int) bool {
// 			return streetNodes[i].Dist < streetNodes[j].Dist
// 		})

// 		nearestStPoint = streetNodes[0].Node
// 		nearestStNodeIdx := streetNodes[0].Idx
// 		secondNearestStPoint = streetNodes[1].Node

// 		// project point ke line segment jalan antara 2 point tadi
// 		projection := ProjectPointToLine(nearestStPoint, secondNearestStPoint, wantToSnap)

// 		projectionLoc := NewLocation(projection.Lat, projection.Lon)

// 		// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
// 		if dir == "f" {
// 			if HaversineDistance(wantToSnapLoc, projectionLoc) < best &&
// 				len(ch.OrigGraph[nearestStNodeIdx].OutEdges) != 0 {
// 				best = HaversineDistance(wantToSnapLoc, projectionLoc)
// 				snappedStNode = nearestStNodeIdx
// 			}
// 		} else {
// 			if HaversineDistance(wantToSnapLoc, projectionLoc) < best &&
// 				len(ch.OrigGraph[nearestStNodeIdx].InEdges) != 0 {
// 				best = HaversineDistance(wantToSnapLoc, projectionLoc)
// 				snappedStNode = nearestStNodeIdx
// 			}
// 		}

// 	}

// 	return snappedStNode, nil
// }
