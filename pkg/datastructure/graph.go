package datastructure

import (
	"lintang/navigatorx/pkg/concurrent"
	"lintang/navigatorx/pkg/geo"

	"github.com/twpayne/go-polyline"
)

type EdgePair struct {
	Weight        float64
	Dist          float64
	ToNodeIDX     int32
	IsShortcut    bool
	EdgeIDx       int32
	Roundabout    bool
	RoadClass     string
	RoadClassLink string
	Lanes         int
}

type Edge struct {
	From          *Node
	To            *Node
	Cost          float64
	StreetName    string
	MaxSpeed      float64
	Roundabout    bool
	RoadClass     string
	RoadClassLink string
	Lanes         int
}

type Node struct {
	Tags         []string
	Out_to       []Edge
	Lat, Lon     float64
	ID           int64
	StreetName   string
	TrafficLight bool
	UsedInRoad   int
}

type CHNode struct {
	OutEdges     []EdgePair
	InEdges      []EdgePair
	Lat          float64
	Lon          float64
	orderPos     int64
	IDx          int32
	StreetName   string
	TrafficLight bool
}

type CHNode2 struct {
	Lat          float64
	Lon          float64
	OrderPos     int64
	IDx          int32
	StreetName   string
	TrafficLight bool
}

type SurakartaWay struct {
	ID                  int32
	CenterLoc           []float64 // [lat, lon]
	Nodes               []CHNode2 // yang bukan intersectionNodes
	IntersectionNodesID []int64
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

type EdgeCH struct {
	EdgeIDx        int32
	Weight         float64
	Dist           float64
	ToNodeIDX      int32
	BaseNodeIDx    int32
	IsShortcut     bool
	RemovedEdgeOne int32
	RemovedEdgeTwo int32
	StreetName     string
	Roundabout     bool
	RoadClass      string
	RoadClassLink  string
	Lanes          int
}

type StreetExtraInfo struct {
	Destination      string
	DestinationRef   string
	MotorwayJunction string
}

func RoadTypeMaxSpeed(roadType string) float64 {
	switch roadType {
	case "motorway":
		return 95
	case "trunk":
		return 85
	case "primary":
		return 75
	case "secondary":
		return 65
	case "tertiary":
		return 50
	case "unclassified":
		return 50
	case "residential":
		return 30
	case "service":
		return 20
	case "motorway_link":
		return 90
	case "trunk_link":
		return 80
	case "primary_link":
		return 70
	case "secondary_link":
		return 60
	case "tertiary_link":
		return 50
	case "living_street":
		return 20
	default:
		return 40
	}
}

type SPSingleResultResult struct {
	Source    int32
	Dest      int32
	Paths     []CHNode2
	EdgePath []EdgeCH
	Dist      float64
	Eta       float64
}

type StateObservationPair struct {
	Observation CHNode2
	State       []State
}

type State struct {
	ID     int
	NodeID int32
	Lat    float64
	Lon    float64
	Dist   float64
	EdgeID int32
}
type SmallWay struct {
	CenterLoc           []float64 // [lat, lon]
	IntersectionNodesID []int64
}

func (s *SmallWay) ToConcurrentWay() concurrent.SmallWay {
	return concurrent.SmallWay{
		CenterLoc:           s.CenterLoc,
		IntersectionNodesID: s.IntersectionNodesID,
	}
}
func (n *CHNode) PathEstimatedCostETA(to CHNode) float64 {

	currLoc := geo.NewLocation(n.Lat, n.Lon)
	toLoc := geo.NewLocation(to.Lat, to.Lon)
	dist := geo.HaversineDistance(currLoc, toLoc) // km

	time := to.OutEdges[0].Weight
	distEdge := to.OutEdges[0].Dist
	speed := (distEdge / time) * 60 / 1000 // km/h

	r := dist / speed // dist = km, speed = km/h
	return r
}

func RenderPath2(path []CHNode2) string {
	s := ""
	coords := make([][]float64, 0)
	for _, p := range path {
		pT := p
		coords = append(coords, []float64{pT.Lat, pT.Lon})
	}
	s = string(polyline.EncodeCoords(coords))
	return s
}
func RenderPath(path []CHNode) string {
	s := ""
	coords := make([][]float64, 0)
	for _, p := range path {
		pT := p
		coords = append(coords, []float64{pT.Lat, pT.Lon})
	}
	s = string(polyline.EncodeCoords(coords))
	return s
}
