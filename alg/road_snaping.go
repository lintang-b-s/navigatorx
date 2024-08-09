package alg

import (
	"math"
	"sort"

	"github.com/dhconnelly/rtreego"
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
	nearestStreets := []NearestStreet{}
	for i, w := range ways {
		street := ways[i]
		if len(street.IntersectionNodesID) < 1 {
			continue
		}

		homeLoc := NewLocation(wantToSnap[0], wantToSnap[1])
		st := NewLocation(w.CenterLoc[0], w.CenterLoc[1])
		nearestStreets = append(nearestStreets, NearestStreet{
			Dist:   HaversineDistance(homeLoc, st),
			Street: &street,
		})
	}

	sort.Slice(nearestStreets, func(i, j int) bool {
		return nearestStreets[i].Dist < nearestStreets[j].Dist
	})

	if len(nearestStreets) >= 15 {
		nearestStreets = nearestStreets[:15]
	}

	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])
	best := 100000000.0
	snappedStNode := int32(0)

	for _, street := range nearestStreets {

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, nodeID := range street.Street.IntersectionNodesID {

			nodeIdx := nodeID
			node := ch.ContractedNodes[nodeIdx]
			nodeLoc := NewLocation(node.Lat, node.Lon)
			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc), int32(nodeIdx)})
		}

		sort.Slice(streetNodes, func(i, j int) bool {
			return streetNodes[i].Dist < streetNodes[j].Dist
		})

		if len(street.Street.IntersectionNodesID) >= 2 {
			nearestStPoint := streetNodes[0].Node
			nearestStNodeIdx := streetNodes[0].Idx
			secondNearestStPoint := streetNodes[1].Node

			// project point ke line segment jalan antara 2 point tadi
			projection := ProjectPointToLineCoord(nearestStPoint, secondNearestStPoint, wantToSnap)

			projectionLoc := NewLocation(projection.Lat, projection.Lon)
			// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
			if HaversineDistance(wantToSnapLoc, projectionLoc) < best {
				best = HaversineDistance(wantToSnapLoc, projectionLoc)
				snappedStNode = nearestStNodeIdx
			}
		} else {
			nearestStPoint := streetNodes[0].Node
			nearestStPointLoc := NewLocation(nearestStPoint.Lat, nearestStPoint.Lon)
			if HaversineDistance(wantToSnapLoc, nearestStPointLoc) < best {
				best = HaversineDistance(wantToSnapLoc, nearestStPointLoc)
				snappedStNode = nearestStPoint.IDx
			}
		}

	}

	return snappedStNode
}

func (ch *ContractedGraph) isPointInsideRoad(street SurakartaWay, p CHNode2) bool {
	pLoc := NewLocation(p.Lat, p.Lon)
	nOne := ch.ContractedNodes[street.IntersectionNodesID[0]]
	nTwo := ch.ContractedNodes[street.IntersectionNodesID[1]]
	stNodeOne := NewLocation(nOne.Lat, nOne.Lon)
	stNodeTwo := NewLocation(nTwo.Lat, nTwo.Lon)
	distOne := HaversineDistance(pLoc, stNodeOne) * 1000       // m
	distTwo := HaversineDistance(pLoc, stNodeTwo) * 1000       // m
	stLength := HaversineDistance(stNodeOne, stNodeTwo) * 1000 // m

	// triangle inequality??
	return math.Abs(distOne+distTwo-stLength) < 0.5
}

func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeH3ForMapMatching(ways []SurakartaWay, wantToSnap []float64) []State {

	sts := []State{}
	nearestStreets := []NearestStreet{}
	for i, w := range ways {
		street := ways[i]

		homeLoc := NewLocation(wantToSnap[0], wantToSnap[1])
		st := NewLocation(w.CenterLoc[0], w.CenterLoc[1])
		nearestStreets = append(nearestStreets, NearestStreet{
			Dist:   HaversineDistance(homeLoc, st),
			Street: &street,
		})
	}

	sort.Slice(nearestStreets, func(i, j int) bool {
		return nearestStreets[i].Dist < nearestStreets[j].Dist
	})

	if len(nearestStreets) >= 10 {
		nearestStreets = nearestStreets[:10]
	}

	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

	for _, st := range nearestStreets {

		street := st.Street

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, node := range street.Nodes {
			nodeIDx := int32(-1)
			val, ok := ch.NodeMapIdx[int64(node.IDx)]
			if ok {
				nodeIDx = val
			}
			nodeLoc := NewLocation(node.Lat, node.Lon)
			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc), nodeIDx})
		}

		if len(streetNodes) < 2 {
			continue
		}

		sort.Slice(streetNodes, func(i, j int) bool {
			return streetNodes[i].Dist < streetNodes[j].Dist
		})

		nearestLoc := NewLocation(streetNodes[0].Node.Lat, streetNodes[0].Node.Lon)
		if HaversineDistance(wantToSnapLoc, nearestLoc)*1000 >= 25 {
			continue
		}
		projection := ProjectPointToLineCoord(streetNodes[0].Node, streetNodes[1].Node, wantToSnap)
		projectionLoc := NewLocation(projection.Lat, projection.Lon)
		sts = append(sts, State{
			NodeID: streetNodes[0].Idx,
			Lat:    projection.Lat,
			Lon:    projection.Lon,
			Dist:   HaversineDistance(wantToSnapLoc, projectionLoc), // pake nearestLoc buat dist nya lumayan bagus
			EdgeID: street.ID,
		})

	}

	for i := len(sts) - 1; i >= 0; i-- {
		if sts[i].Dist*1000 >= 25 {
			sts[i] = sts[len(sts)-1]
			sts = sts[:len(sts)-1]
		}
	}

	// bagusan pake rtree & projection di lat & lon nya
	// max dist 25 paling bagus

	return sts
}

// func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (states []State, err error) {
// 	sts := []State{}

// 	wantToSnap := rtreego.Point{lat, lon}
// 	stNeighbors := ch.Rtree.StRtree.NearestNeighbors(4, wantToSnap)

// 	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

// 	for _, st := range stNeighbors {

// 		street := st.(*StreetRect).Street

// 		// mencari 2 point dijalan yg paling dekat dg gps
// 		streetNodes := []NodePoint{}
// 		for _, nodeID := range street.NodesID {
// 			nodeIdx := nodeID
// 			node := ch.ContractedNodes[nodeIdx]
// 			nodeLoc := NewLocation(node.Lat, node.Lon)
// 			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc), int32(nodeIdx)})
// 		}

// 		sort.Slice(streetNodes, func(i, j int) bool {
// 			return streetNodes[i].Dist < streetNodes[j].Dist
// 		})

// 		nearestLoc := NewLocation(streetNodes[0].Node.Lat, streetNodes[0].Node.Lon)
// 		if HaversineDistance(wantToSnapLoc, nearestLoc)*1000 >= 25 {
// 			continue
// 		}
// 		projection := ProjectPointToLineCoord(streetNodes[0].Node, streetNodes[1].Node, wantToSnap)
// 		projectionLoc := NewLocation(projection.Lat, projection.Lon)
// 		sts = append(sts, State{
// 			NodeID:    streetNodes[0].Idx,
// 			Lat:       projection.Lat,
// 			Lon:       projection.Lon,
// 			Dist:      HaversineDistance(wantToSnapLoc, projectionLoc), // pake nearestLoc buat dist nya lumayan bagus
// 			EdgeBound: street.Bound,
// 		})

// 	}

// 	for i := len(sts) - 1; i >= 0; i-- {
// 		if sts[i].Dist*1000 >= 25 {
// 			sts[i] = sts[len(sts)-1]
// 			sts = sts[:len(sts)-1]
// 		}
// 	}

// 	sort.Slice(sts, func(i, j int) bool {
// 		return sts[i].Dist < sts[j].Dist
// 	})

// 	return sts, nil
// }

func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNodeIdx int32, err error) {
	wantToSnap := rtreego.Point{lat, lon}
	stNeighbors := ch.Rtree.StRtree.NearestNeighbors(5, wantToSnap)

	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

	snappedStNode := int32(0)
	best := 100000000.0

	// snap point ke  node jalan terdekat/posisi location seharusnya
	for _, st := range stNeighbors {
		if len(st.(*StreetRect).Street.IntersectionNodesID) < 1 {
			continue
		}

		street := st.(*StreetRect).Street

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, nodeID := range street.IntersectionNodesID {
			nodeIdx := nodeID
			node := ch.ContractedNodes[nodeIdx]
			nodeLoc := NewLocation(node.Lat, node.Lon)
			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc), int32(nodeIdx)})
		}

		sort.Slice(streetNodes, func(i, j int) bool {
			return streetNodes[i].Dist < streetNodes[j].Dist
		})

		if len(street.IntersectionNodesID) >= 2 {
			nearestStPoint := streetNodes[0].Node
			nearestStNodeIdx := streetNodes[0].Idx
			secondNearestStPoint := streetNodes[1].Node

			// project point ke line segment jalan antara 2 point tadi
			projection := ProjectPointToLineCoord(nearestStPoint, secondNearestStPoint, wantToSnap)

			projectionLoc := NewLocation(projection.Lat, projection.Lon)
			// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
			if HaversineDistance(wantToSnapLoc, projectionLoc) < best {
				best = HaversineDistance(wantToSnapLoc, projectionLoc)
				snappedStNode = nearestStNodeIdx
			}
		} else {
			nearestStPoint := streetNodes[0].Node
			nearestStPointLoc := NewLocation(nearestStPoint.Lat, nearestStPoint.Lon)
			if HaversineDistance(wantToSnapLoc, nearestStPointLoc) < best {
				best = HaversineDistance(wantToSnapLoc, nearestStPointLoc)
				snappedStNode = nearestStPoint.IDx
			}
		}
	}

	return snappedStNode, nil
}

// for _, street := range nearestStreets {
// 	if len(street.Street.Nodes) < 2 {
// 		continue
// 	}
// 	notIntersectionNodes := []NodePoint{}
// 	for _, node := range street.Street.Nodes {
// 		nodeLoc := NewLocation(node.Lat, node.Lon)
// 		notIntersectionNodes = append(notIntersectionNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc), -1})
// 	}

// 	sort.Slice(notIntersectionNodes, func(i, j int) bool {
// 		return notIntersectionNodes[i].Dist < notIntersectionNodes[j].Dist
// 	})
// 	nearestNotIntersectionNode := notIntersectionNodes[0].Node
// 	secondNearestNotIntersectionNode := notIntersectionNodes[1].Node
// 	projectionNotIntersection := ProjectPointToLineCoord(nearestNotIntersectionNode, secondNearestNotIntersectionNode, wantToSnap)
// 	projectionNotIntersectionLoc := NewLocation(projectionNotIntersection.Lat, projectionNotIntersection.Lon)

// 	if HaversineDistance(wantToSnapLoc, projectionNotIntersectionLoc) < bestNotIntersection {

// 		bestNotIntersection = HaversineDistance(wantToSnapLoc, projectionNotIntersectionLoc)
// 		snappedNotIntersectionNodes = notIntersectionNodes[0].Node

// 	}
// }

// fmt.Println("best: ", snappedNotIntersectionNodes)
// isNotIntersectionNodeInsideBestSt := ch.isPointInsideRoad(bestStreet, snappedNotIntersectionNodes)
// fmt.Println(isNotIntersectionNodeInsideBestSt)

// snappedNotIntersectionNodes := CHNode2{}
// bestNotIntersection := 100000000.0
// bestStreet := SurakartaWay{}
