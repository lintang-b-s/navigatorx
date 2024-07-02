package alg

import "github.com/twpayne/go-polyline"

type Edge struct {
	From *Node
	To   *Node
	Cost float64
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

	r := float64(absLat + absLon)
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
