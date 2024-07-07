package service

import (
	"context"
	"lintang/coba_osm/alg"
	"lintang/coba_osm/domain"
	"lintang/coba_osm/util"
	"sync"
)

type NavigationService struct {
}

func NewNavigationService() *NavigationService {
	return &NavigationService{}
}

func (uc *NavigationService) ShortestPathETA(ctx context.Context, srcLat, srcLon float64,
	dstLat float64, dstLon float64) (string, float64, []alg.Navigation, bool, []alg.Coordinate, float64, error) {

	from := &alg.Node{
		Lat: util.TruncateFloat64(srcLat, 6),
		Lon: util.TruncateFloat64(srcLon, 6),
	}
	to := &alg.Node{
		Lat: util.TruncateFloat64(dstLat, 6),
		Lon: util.TruncateFloat64(dstLon, 6),
	}

	var err error
	fromSurakartaNode, err := alg.SnapLocationToRoadNetworkNodeRtree(from.Lat, from.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, nil
	}
	toSurakartaNode, err := alg.SnapLocationToRoadNetworkNodeRtree(to.Lat, to.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, nil
	}

	if fromSurakartaNode == nil || toSurakartaNode == nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, nil
	}

	p, eta, found, dist := alg.AStarETA(fromSurakartaNode, toSurakartaNode)
	// eta satuannya minute
	// dist := 0
	if !found {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, domain.WrapErrorf(err, domain.ErrNotFound, "sorry!! lokasi yang anda masukkan tidak tercakup di peta saya :(")
	}
	var route []alg.Coordinate = make([]alg.Coordinate, 0)
	for i := range p {
		pathN := *p[len(p)-1-i].(*alg.Node)

		route = append(route, alg.Coordinate{
			Lat: pathN.Lat,
			Lon: pathN.Lon,
		})
	}

	n, err := alg.CreateTurnByTurnNavigation(reverse(p))
	if err != nil {
		return alg.RenderPath(p), dist, n, found, route, eta, nil
	}

	return alg.RenderPath(p), dist, n, found, route, eta, nil
}

type ShortestPathResult struct {
	Paths []alg.Pather
	ETA   float64
	Found bool
	Dist  float64
	Index int
}

func (uc *NavigationService) ShortestPathAlternativeStreetETA(ctx context.Context, srcLat, srcLon float64,
	alternativeStreetLat float64, alternativeStreetLon float64,
	dstLat float64, dstLon float64) (string, float64, []alg.Navigation, bool, []alg.Coordinate, float64, error) {

	from := &alg.Node{
		Lat: util.TruncateFloat64(srcLat, 6),
		Lon: util.TruncateFloat64(srcLon, 6),
	}

	alternativeStreet := &alg.Node{
		Lat: util.TruncateFloat64(alternativeStreetLat, 6),
		Lon: util.TruncateFloat64(alternativeStreetLon, 6),
	}

	to := &alg.Node{
		Lat: util.TruncateFloat64(dstLat, 6),
		Lon: util.TruncateFloat64(dstLon, 6),
	}

	var err error
	fromSurakartaNode, err := alg.SnapLocationToRoadNetworkNodeRtree(from.Lat, from.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, nil
	}

	alternativeStreetSurakartaNode, err := alg.SnapLocationToRoadNetworkNodeRtree(alternativeStreet.Lat, alternativeStreet.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, nil
	}

	toSurakartaNode, err := alg.SnapLocationToRoadNetworkNodeRtree(to.Lat, to.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, nil
	}

	if fromSurakartaNode == nil || toSurakartaNode == nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, nil
	}

	// concurrently find the shortest path dari fromSurakartaNode ke alternativeStreetSurakartaNode
	// dan dari alternativeStreetSurakartaNode ke toSurakartaNode

	var wg sync.WaitGroup

	paths := make([]ShortestPathResult, 2)

	pathChan := make(chan ShortestPathResult, 2)
	wg.Add(1)
	wg.Add(1)
	go func() {

		defer wg.Done()
		p, eta, found, dist := alg.AStarETA(fromSurakartaNode, alternativeStreetSurakartaNode)
		pathChan <- ShortestPathResult{
			Paths: p,
			ETA:   eta,
			Found: found,
			Dist:  dist,
			Index: 0,
		}
	}()

	go func() {

		defer wg.Done()
		p, eta, found, dist := alg.AStarETA(alternativeStreetSurakartaNode, toSurakartaNode)
		pathChan <- ShortestPathResult{
			Paths: p,
			ETA:   eta,
			Found: found,
			Dist:  dist,
			Index: 1,
		}
	}()

	i := 0
	for p := range pathChan {
		if p.Index == 0 {
			paths[0] = p
		} else {
			paths[1] = p
		}
		i++
		if i == 2 {
			break
		}
	}

	wg.Wait()

	concatedPaths := []alg.Pather{}
	paths[0].Paths = reverse(paths[0].Paths[:len(paths[0].Paths)-1]) // exclude start node dari paths[1]
	paths[1].Paths = reverse(paths[1].Paths)
	concatedPaths = append(concatedPaths, paths[0].Paths...)
	concatedPaths = append(concatedPaths, paths[1].Paths...)

	eta := paths[0].ETA + paths[1].ETA
	dist := paths[0].Dist + paths[1].Dist
	found := paths[0].Found && paths[1].Found

	// eta satuannya minute
	// dist := 0
	if !found {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, domain.WrapErrorf(err, domain.ErrNotFound, "sorry!! lokasi yang anda masukkan tidak tercakup di peta saya :(")
	}
	var route []alg.Coordinate = make([]alg.Coordinate, 0)
	for i := range concatedPaths {
		pathN := *concatedPaths[len(concatedPaths)-1-i].(*alg.Node)

		route = append(route, alg.Coordinate{
			Lat: pathN.Lat,
			Lon: pathN.Lon,
		})
	}

	n, err := alg.CreateTurnByTurnNavigation(concatedPaths)
	if err != nil {
		return alg.RenderPath(concatedPaths), dist, n, found, route, eta, nil
	}

	return alg.RenderPath(concatedPaths), dist, n, found, route, eta, nil
}

func reverse(p []alg.Pather) []alg.Pather {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
	return p
}
