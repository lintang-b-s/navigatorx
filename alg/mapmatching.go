package alg

import (
	"sort"

	"github.com/dhconnelly/rtreego"
)

type NodePoint struct {
	Node *Node
	Dist float64
}

func SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNode *Node, err error) {
	wantToSnap := rtreego.Point{lat, lon}
	stNeighbors := StRTree.NearestNeighbors(3, wantToSnap)

	wantToSnapLoc := NewLocation(wantToSnap[0], wantToSnap[1])

	snappedStNode := &Node{}
	best := 100000000.0

	// snap point ke  node jalan terdekat/posisi location seharusnya
	for _, st := range stNeighbors {

		street := st.(*StreetRect).Street
		nearestStPoint := street.Nodes[0]       // node di jalan yg paling dekat dg gps
		secondNearestStPoint := street.Nodes[0] // node di jalan yang paling dekat kedua dg gps

		// mencari 2 point dijalan yg paling dekat dg gps
		streetNodes := []NodePoint{}
		for _, node := range street.Nodes {
			nodeLoc := NewLocation(node.Lat, node.Lon)
			streetNodes = append(streetNodes, NodePoint{node, HaversineDistance(wantToSnapLoc, nodeLoc)})
		}

		sort.Slice(streetNodes, func(i, j int) bool {
			return streetNodes[i].Dist < streetNodes[j].Dist
		})

		nearestStPoint = streetNodes[0].Node
		secondNearestStPoint = streetNodes[1].Node

		// project point ke line segment jalan antara 2 point tadi
		projection := ProjectPointToLine(*nearestStPoint, *secondNearestStPoint, wantToSnap)

		projectionLoc := NewLocation(projection.Lat, projection.Lon)

		// ambil streetNode yang jarak antara hasil projection dg lokasi gps  paling kecil
		if HaversineDistance(wantToSnapLoc, projectionLoc) < best {
			best = HaversineDistance(wantToSnapLoc, projectionLoc)
			snappedStNode = nearestStPoint
		}
	}

	return snappedStNode, nil
}
