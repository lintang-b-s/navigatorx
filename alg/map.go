package alg

import (
	"lintang/coba_osm/util"
	"sort"

	"github.com/paulmach/osm"
)

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

var SurakartaNodeMap = make(map[int64]*Node)

var SurakartaGraphData = SurakartaGraph{
	Nodes:   make([]*Node, 0),
	NodeIdx: make(map[int64]int64),
	Counter: 0,
	Ways:    make([]SurakartaWay, 0),
}

type Bound struct {
	MinLat float64 `json:"minlat"`
	MaxLat float64 `json:"maxlat"`
	MinLon float64 `json:"minlon"`
	MaxLon float64 `json:"maxlon"`
}

type SurakartaWay struct {
	ID        int64
	Nodes     []*Node
	Bound     Bound
	CenterLoc []float64 // [lat, lon]
}

type SurakartaGraph struct {
	Nodes   []*Node
	NodeIdx map[int64]int64
	Counter int64
	Ways    []SurakartaWay
}

func InitGraph(ways []*osm.Way) {

	for _, way := range ways {
		sWay := SurakartaWay{
			ID:    int64(way.ID),
			Nodes: make([]*Node, 0),
		}

		streetNodeLats := []float64{}
		streetNodeLon := []float64{}

		// creategraph node
		for i := 0; i < len(way.Nodes)-1; i++ {
			fromN := way.Nodes[i]

			from := &Node{
				Lat: util.RoundFloat(fromN.Lat, 6),
				Lon: util.RoundFloat(fromN.Lon, 6),
				ID:  int64(fromN.ID),
			}

			toN := way.Nodes[i+1]
			to := &Node{
				Lat: util.RoundFloat(toN.Lat, 6),
				Lon: util.RoundFloat(toN.Lon, 6),
				ID:  int64(toN.ID),
			}

			if fromRealNode, ok := SurakartaNodeMap[from.ID]; ok {
				from = fromRealNode
			} else {
				SurakartaNodeMap[from.ID] = from
			}
			if toRealNode, ok := SurakartaNodeMap[to.ID]; ok {
				to = toRealNode
			} else {
				SurakartaNodeMap[to.ID] = to
			}

			edge := Edge{
				From: from,
				To:   to,
				Cost: EuclideanDistance(from, to),
			}
			from.Out_to = append(from.Out_to, edge)

			reverseEdge := Edge{
				From: to,
				To:   from,
				Cost: EuclideanDistance(from, to),
			}

			to.Out_to = append(to.Out_to, reverseEdge)

			if _, ok := SurakartaGraphData.NodeIdx[from.ID]; ok {
				fromIdx := SurakartaGraphData.NodeIdx[from.ID]
				SurakartaGraphData.Nodes[fromIdx] = from
			} else {
				SurakartaGraphData.NodeIdx[from.ID] = SurakartaGraphData.Counter // save index node saat ini
				SurakartaGraphData.Nodes = append(SurakartaGraphData.Nodes, from)
				SurakartaGraphData.Counter++
			}
			if _, ok := SurakartaGraphData.NodeIdx[to.ID]; ok {
				toIdx := SurakartaGraphData.NodeIdx[to.ID]
				SurakartaGraphData.Nodes[toIdx] = to
			} else {
				SurakartaGraphData.NodeIdx[to.ID] = SurakartaGraphData.Counter
				SurakartaGraphData.Nodes = append(SurakartaGraphData.Nodes, to)
				SurakartaGraphData.Counter++
			}

			// add node ke surakartaway
			sWay.Nodes = append(sWay.Nodes, from)
			// append node lat & node lon
			streetNodeLats = append(streetNodeLats, from.Lat)
			streetNodeLon = append(streetNodeLon, from.Lon)
			if i == len(way.Nodes)-2 {
				sWay.Nodes = append(sWay.Nodes, to)
				streetNodeLats = append(streetNodeLats, to.Lat)
				streetNodeLon = append(streetNodeLon, to.Lon)
			}
		}
		sort.Sort(sort.Float64Slice(streetNodeLats))
		sort.Sort(sort.Float64Slice(streetNodeLon))
		sWay.Bound.MinLat = streetNodeLats[0]
		sWay.Bound.MaxLat = streetNodeLats[len(streetNodeLats)-1]
		sWay.Bound.MinLon = streetNodeLon[0]
		sWay.Bound.MaxLon = streetNodeLon[len(streetNodeLon)-1]
		sWay.CenterLoc = []float64{(sWay.Bound.MinLat + sWay.Bound.MaxLat) / 2, (sWay.Bound.MinLon + sWay.Bound.MaxLon) / 2}

		SurakartaGraphData.Ways = append(SurakartaGraphData.Ways, sWay)
	}
}
