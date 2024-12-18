package contractor

import (
	"lintang/navigatorx/pkg/datastructure"
	"lintang/navigatorx/pkg/geo"
	"lintang/navigatorx/pkg/guidance"
	"sort"
)

// urutin field struct descending by size , biar makin kecil heap size nya
type NodePoint struct {
	Node datastructure.CHNode2
	Dist float64
	Idx  int32
}

type SmallWay struct {
	CenterLoc           []float64 // [lat, lon]
	IntersectionNodesID []int64
}

type NearestStreet struct {
	Dist   float64
	Street *SmallWay
}

func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeH3(ways []datastructure.SmallWay, wantToSnap []float64) int32 {
	nearestStreets := []NearestStreet{}
	for i, w := range ways {
		street := ways[i]
		if len(street.IntersectionNodesID) < 1 {
			continue
		}

		homeLoc := geo.NewLocation(wantToSnap[0], wantToSnap[1])
		st := geo.NewLocation(w.CenterLoc[0], w.CenterLoc[1])
		nearestStreets = append(nearestStreets, NearestStreet{
			Dist: geo.HaversineDistance(homeLoc, st),
			Street: &SmallWay{
				CenterLoc:           w.CenterLoc,
				IntersectionNodesID: w.IntersectionNodesID,
			},
		})
	}

	sort.Slice(nearestStreets, func(i, j int) bool {
		return nearestStreets[i].Dist < nearestStreets[j].Dist
	})

	if len(nearestStreets) >= 7 {
		nearestStreets = nearestStreets[:7]
	}

	wantToSnapLoc := geo.NewLocation(wantToSnap[0], wantToSnap[1])
	best := 100000000.0
	snappedStNode := int32(0)

	for _, street := range nearestStreets {

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, nodeID := range street.Street.IntersectionNodesID {

			nodeIdx := nodeID
			node := ch.ContractedNodes[nodeIdx]
			nodeLoc := geo.NewLocation(node.Lat, node.Lon)
			cNode := datastructure.CHNode2{
				Lat:          node.Lat,
				Lon:          node.Lon,
				OrderPos:     node.OrderPos,
				IDx:          node.IDx,
				StreetName:   node.StreetName,
				TrafficLight: node.TrafficLight,
			}
			streetNodes = append(streetNodes, NodePoint{cNode, geo.HaversineDistance(wantToSnapLoc, nodeLoc), int32(nodeIdx)})
		}

		sort.Slice(streetNodes, func(i, j int) bool {
			return streetNodes[i].Dist < streetNodes[j].Dist
		})

		if len(street.Street.IntersectionNodesID) >= 2 {
			nearestStPoint := streetNodes[0].Node
			nearestStNodeIdx := streetNodes[0].Idx
			secondNearestStPoint := streetNodes[1].Node

			// project point ke line segment jalan antara 2 point tadi
			nearestStPointGuidance := datastructure.CHNode2{
				Lat:          nearestStPoint.Lat,
				Lon:          nearestStPoint.Lon,
				OrderPos:     nearestStPoint.OrderPos,
				IDx:          nearestStNodeIdx,
				StreetName:   nearestStPoint.StreetName,
				TrafficLight: nearestStPoint.TrafficLight,
			}
			secondNearestStPointGuidance := datastructure.CHNode2{
				Lat:          secondNearestStPoint.Lat,
				Lon:          secondNearestStPoint.Lon,
				OrderPos:     secondNearestStPoint.OrderPos,
				IDx:          secondNearestStPoint.IDx,
				StreetName:   secondNearestStPoint.StreetName,
				TrafficLight: secondNearestStPoint.TrafficLight,
			}
			projection := guidance.ProjectPointToLineCoord(nearestStPointGuidance, secondNearestStPointGuidance, wantToSnap)

			projectionLoc := geo.NewLocation(projection.Lat, projection.Lon)
			// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
			if geo.HaversineDistance(wantToSnapLoc, projectionLoc) < best {
				best = geo.HaversineDistance(wantToSnapLoc, projectionLoc)
				snappedStNode = nearestStNodeIdx
			}
		} else {
			nearestStPoint := streetNodes[0].Node
			nearestStPointLoc := geo.NewLocation(nearestStPoint.Lat, nearestStPoint.Lon)
			if geo.HaversineDistance(wantToSnapLoc, nearestStPointLoc) < best {
				best = geo.HaversineDistance(wantToSnapLoc, nearestStPointLoc)
				snappedStNode = nearestStPoint.IDx
			}
		}

	}

	return snappedStNode
}

type State struct {
	ID     int
	NodeID int32
	Lat    float64
	Lon    float64
	Dist   float64
	EdgeID int32
}

func (ch *ContractedGraph) SnapLocationToRoadNetworkNodeH3ForMapMatching(ways []datastructure.SmallWay, wantToSnap []float64) []datastructure.State {

	sts := []State{}
	nearestStreets := []NearestStreet{}
	for _, w := range ways {
		// street := ways[i]

		homeLoc := geo.NewLocation(wantToSnap[0], wantToSnap[1])
		st := geo.NewLocation(w.CenterLoc[0], w.CenterLoc[1])
		nearestStreets = append(nearestStreets, NearestStreet{
			Dist: geo.HaversineDistance(homeLoc, st),
			Street: &SmallWay{
				CenterLoc:           w.CenterLoc,
				IntersectionNodesID: w.IntersectionNodesID,
			},
		})
	}

	sort.Slice(nearestStreets, func(i, j int) bool {
		return nearestStreets[i].Dist < nearestStreets[j].Dist
	})

	if len(nearestStreets) >= 7 {
		nearestStreets = nearestStreets[:7]
	}

	wantToSnapLoc := geo.NewLocation(wantToSnap[0], wantToSnap[1])

	for idx, st := range nearestStreets {

		street := st.Street

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		if len(street.IntersectionNodesID) == 0 {
			continue
		}
		for _, nodeID := range street.IntersectionNodesID {

			nodeIdx := nodeID
			node := ch.ContractedNodes[nodeIdx]
			nodeLoc := geo.NewLocation(node.Lat, node.Lon)
			cNode := datastructure.CHNode2{
				Lat:          node.Lat,
				Lon:          node.Lon,
				OrderPos:     node.OrderPos,
				IDx:          node.IDx,
				StreetName:   node.StreetName,
				TrafficLight: node.TrafficLight,
			}
			streetNodes = append(streetNodes, NodePoint{cNode, geo.HaversineDistance(wantToSnapLoc, nodeLoc), int32(nodeIdx)})
		}


		nearestLoc := geo.NewLocation(streetNodes[0].Node.Lat, streetNodes[0].Node.Lon)
		if geo.HaversineDistance(wantToSnapLoc, nearestLoc)*1000 >= 25 {
			continue
		}
		streetOne := datastructure.CHNode2{
			Lat:          streetNodes[0].Node.Lat,
			Lon:          streetNodes[0].Node.Lon,
			OrderPos:     streetNodes[0].Node.OrderPos,
			IDx:          streetNodes[0].Node.IDx,
			StreetName:   streetNodes[0].Node.StreetName,
			TrafficLight: streetNodes[0].Node.TrafficLight,
		}

		streetTwo := datastructure.CHNode2{
			Lat:          streetNodes[1].Node.Lat,
			Lon:          streetNodes[1].Node.Lon,
			OrderPos:     streetNodes[1].Node.OrderPos,
			IDx:          streetNodes[1].Node.IDx,
			StreetName:   streetNodes[1].Node.StreetName,
			TrafficLight: streetNodes[1].Node.TrafficLight,
		}
		projection := guidance.ProjectPointToLineCoord(streetOne, streetTwo, wantToSnap)
		projectionLoc := geo.NewLocation(projection.Lat, projection.Lon)
		sts = append(sts, State{
			NodeID: streetNodes[0].Idx,
			Lat:    projection.Lat,
			Lon:    projection.Lon,
			Dist:   geo.HaversineDistance(wantToSnapLoc, projectionLoc), // pake nearestLoc buat dist nya lumayan bagus
			EdgeID: int32(idx),
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
	stsData := make([]datastructure.State, len(sts))
	for i, st := range sts {
		stsData[i] = datastructure.State{
			ID:     st.ID,
			NodeID: st.NodeID,
			Lat:    st.Lat,
			Lon:    st.Lon,
			Dist:   st.Dist,
			EdgeID: st.EdgeID,
		}
	}

	return stsData
}
