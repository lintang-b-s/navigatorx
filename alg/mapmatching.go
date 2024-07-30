package alg

import (
	"sort"
)

// urutin field struct descending by size , biar makin kecil heap size nya
type NodePoint struct {
	Node CHNode2
	Dist float64
	Idx  int32
}
type NearestStreet struct {
	Dist   float64
	Street *SurakartaWay
}

func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeH3(ways []SurakartaWay, wantToSnap []float64) int32 {
	nearest := []NearestStreet{}
	for i, w := range ways {
		street := ways[i]

		homeLoc := NewLocation(wantToSnap[0], wantToSnap[1])
		st := NewLocation(float64(w.CenterLoc[0]), float64(w.CenterLoc[1]))
		nearest = append(nearest, NearestStreet{
			Dist:   HaversineDistance(homeLoc, st),
			Street: &street,
		})
	}

	sort.Slice(nearest, func(i, j int) bool {
		return nearest[i].Dist < nearest[j].Dist
	})

	if len(nearest) >= 3 {
		nearest = nearest[:3]
	}
	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])
	best := 100000000.0
	snappedStNode := int32(0)

	for _, street := range nearest {
		// nearest := street.Street.NodesID[0]

		nearest := ch.ContractedNodes[street.Street.NodesID[0]]
		nearestStPoint := nearest       // node di jalan yg paling dekat dg gps
		secondNearestStPoint := nearest // node di jalan yang paling dekat kedua dg gps

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, nodeID := range street.Street.NodesID {
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

	return snappedStNode
}

// func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNodeIdx int32, err error) {
// 	wantToSnap := rtreego.Point{lat, lon}
// 	stNeighbors := ch.Rtree.StRtree.NearestNeighbors(4, wantToSnap)

// 	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

// 	snappedStNode := int32(0)
// 	best := 100000000.0

// 	// snap point ke  node jalan terdekat/posisi location seharusnya
// 	for _, st := range stNeighbors {

// 		street := st.(*StreetRect).Street
// 		nearest := ch.ContractedNodes[street.NodesID[0]]
// 		nearestStPoint := nearest       // node di jalan yg paling dekat dg gps
// 		secondNearestStPoint := nearest // node di jalan yang paling dekat kedua dg gps

// 		// mencari 2 point dijalan yg paling dekat dg gps
// 		streetNodes := []NodePoint{}
// 		for _, nodeID := range street.NodesID {
// 			nodeIdx := nodeID
// 			node := ch.ContractedNodes[nodeIdx]
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
// 		if HaversineDistance(wantToSnapLoc, projectionLoc) < best {
// 			best = HaversineDistance(wantToSnapLoc, projectionLoc)
// 			snappedStNode = nearestStNodeIdx
// 		}
// 	}

// 	return snappedStNode, nil
// }
