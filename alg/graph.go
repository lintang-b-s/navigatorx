package alg

import (
	"github.com/twpayne/go-polyline"
)

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
