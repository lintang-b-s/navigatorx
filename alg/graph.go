package alg

import (
	"math"

	"github.com/twpayne/go-polyline"
)

type Edge struct {
	From     *Node
	To       *Node
	Cost     float64
	MaxSpeed float64
}

type Node struct {
	Lat, Lon float64

	Out_to []Edge

	ID   int64
	tags []string
}

func (n *Node) PathNeighbors() []Pather {
	neighbors := []Pather{}

	for _, e := range n.Out_to {
		neighbors = append(neighbors, Pather(e.To))
	}

	return neighbors
}



func (n *Node) PathNeighborCost(to Pather) float64 {

	for _, e := range n.Out_to {
		if Pather(e.To) == to {
			return e.Cost
		}
	}

	return 10000000
}

func (n *Node) PathEstimatedCost(to Pather) float64 {

	toN := to.(*Node)
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
	r := math.Sqrt(absLatSq + absLonSq)
	return r
}

func (n *Node) PathNeighborCostETA(to Pather) float64 {

	for _, e := range n.Out_to {
		if Pather(e.To) == to {
			maxSpeed := e.MaxSpeed * 1000 / 60 // m/min
			return (e.Cost * 100000) / maxSpeed // minute
		}
	}

	return 100000000000
}

func (n *Node) PathEstimatedCostETA(to Pather) float64 {

	toN := to.(*Node)
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
	maxSpeed := 90.0 * 1000.0 / 60.0 // m/min
	r := math.Sqrt(absLatSq+absLonSq) * 100000 / maxSpeed // * 100000 -> meter
	return r
}

func RenderPath(path []Pather) string {
	s := ""
	coords := make([][]float64, 0)
	for _, p := range path {
		pT := p.(*Node)
		// s = fmt.Sprint(pT.ID) + " " + s
		coords = append(coords, []float64{pT.Lat, pT.Lon})
		// s = fmt.Sprint(idx) + " " + fmt.Sprintf("%f", pT.Lat) + ", " + fmt.Sprintf("%f", pT.Lon) + " \n" + s
	}
	s = string(polyline.EncodeCoords(coords))
	return s
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
	case "pedestrian":
		return 20
	case "track":
		return 20
	case "bus_guideway":
		return 20
	case "escape":
		return 20
	case "services":
		return 20
	case "raceway":
		return 50
	default:
		return 40
	}
}
