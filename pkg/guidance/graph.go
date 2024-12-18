package guidance

type EdgePair struct {
	Weight         float64
	Dist           float64
	ToNodeIDX      int32
	IsShortcut     bool
	EdgeIDx        int32
	RemovedEdgeOne *EdgePair
	RemovedEdgeTwo *EdgePair
}

type Edge struct {
	From     *Node
	To       *Node
	Cost     float64
	MaxSpeed float64
}

type Node struct {
	tags         []string
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
