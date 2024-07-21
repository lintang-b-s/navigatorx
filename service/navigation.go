package service

import (
	"context"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/domain"
	"lintang/navigatorx/util"
	"sync"
)

type ContractedGraph interface {
	SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNodeIdx int32, err error)
	SnapLocationToRoadNetworkNodeRtreeCH(lat, lon float64, dir string) (snappedRoadNodeIdx int32, err error)
	ShortestPathBiDijkstra(from, to int32) ([]alg.CHNode, float64, float64)
	AStarCH(from, to int32) (pathN []alg.CHNode, path string, eta float64, found bool, dist float64)
	IsChReady() bool
}

type NavigationService struct {
	CH ContractedGraph
}

func NewNavigationService(contractedGraph ContractedGraph) *NavigationService {
	return &NavigationService{CH: contractedGraph}
}
func (uc *NavigationService) ShortestPathETA(ctx context.Context, srcLat, srcLon float64,
	dstLat float64, dstLon float64) (string, float64, []alg.Navigation, bool, []alg.Coordinate, float64, bool, error) {

	from := &alg.Node{
		Lat: util.TruncateFloat64(srcLat, 6),
		Lon: util.TruncateFloat64(srcLon, 6),
	}
	to := &alg.Node{
		Lat: util.TruncateFloat64(dstLat, 6),
		Lon: util.TruncateFloat64(dstLon, 6),
	}

	var err error
	fromSurakartaNode, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(from.Lat, from.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}
	toSurakartaNode, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(to.Lat, to.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}

	var pN = []alg.CHNode{}
	var p string
	var eta float64
	var found bool
	var dist float64
	var isCH bool
	if uc.CH.IsChReady() {
		pN, eta, dist = uc.CH.ShortestPathBiDijkstra(fromSurakartaNode, toSurakartaNode)
		p = alg.RenderPath(pN)
		if eta != -1 {
			found = true
		}
		isCH = true
	} else {
		pN, p, eta, found, dist = uc.CH.AStarCH(fromSurakartaNode, toSurakartaNode)
	}

	if !found {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, domain.WrapErrorf(err, domain.ErrNotFound, "sorry!! lokasi yang anda masukkan tidak tercakup di peta saya :(")
	}
	var route []alg.Coordinate = make([]alg.Coordinate, 0)

	n, _ := alg.CreateTurnByTurnNavigation(pN)

	return p, dist, n, found, route, eta, isCH, nil
}

type ShortestPathResult struct {
	Paths []alg.CHNode
	ETA   float64
	Found bool
	Dist  float64
	Index int
	IsCH  bool
}

func (uc *NavigationService) ShortestPathAlternativeStreetETA(ctx context.Context, srcLat, srcLon float64,
	alternativeStreetLat float64, alternativeStreetLon float64,
	dstLat float64, dstLon float64) (string, float64, []alg.Navigation, bool, []alg.Coordinate, float64, bool, error) {

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
	fromSurakartaNode, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(from.Lat, from.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}

	alternativeStreetSurakartaNode, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(alternativeStreet.Lat, alternativeStreet.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}

	toSurakartaNode, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(to.Lat, to.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
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
		var pN = []alg.CHNode{}
		var eta float64
		var found bool
		var dist float64
		var isCH bool

		if uc.CH.IsChReady() {
			pN, eta, dist = uc.CH.ShortestPathBiDijkstra(fromSurakartaNode, alternativeStreetSurakartaNode)
			if eta != -1 {
				found = true
			}
			isCH = true
		} else {
			pN, _, eta, found, dist = uc.CH.AStarCH(fromSurakartaNode, alternativeStreetSurakartaNode)
		}
		pathChan <- ShortestPathResult{
			Paths: pN,
			ETA:   eta,
			Found: found,
			Dist:  dist,
			Index: 0,
			IsCH:  isCH,
		}
	}()

	go func() {

		defer wg.Done()
		var pN = []alg.CHNode{}
		var eta float64
		var found bool
		var dist float64
		var isCH bool

		if uc.CH.IsChReady() {
			pN, eta, dist = uc.CH.ShortestPathBiDijkstra(alternativeStreetSurakartaNode, toSurakartaNode)
			if eta != -1 {
				found = true
			}
			isCH = true
		} else {
			pN, _, eta, found, dist = uc.CH.AStarCH(alternativeStreetSurakartaNode, toSurakartaNode)

		}
		pathChan <- ShortestPathResult{
			Paths: pN,
			ETA:   eta,
			Found: found,
			Dist:  dist,
			Index: 1,
			IsCH:  isCH,
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
	close(pathChan)

	concatedPaths := []alg.CHNode{}
	paths[0].Paths = paths[0].Paths[:len(paths[0].Paths)-1] // exclude start node dari paths[1]
	concatedPaths = append(concatedPaths, paths[0].Paths...)
	concatedPaths = append(concatedPaths, paths[1].Paths...)

	eta := paths[0].ETA + paths[1].ETA
	dist := paths[0].Dist + paths[1].Dist
	found := paths[0].Found && paths[1].Found
	isCH := paths[0].IsCH
	// eta satuannya minute
	// dist := 0
	if !found {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, domain.WrapErrorf(err, domain.ErrNotFound, "sorry!! lokasi yang anda masukkan tidak tercakup di peta saya :(")
	}
	var route []alg.Coordinate = make([]alg.Coordinate, 0)

	n, err := alg.CreateTurnByTurnNavigation(concatedPaths)
	if err != nil {
		return alg.RenderPath(concatedPaths), dist, n, found, route, eta, isCH, nil
	}

	return alg.RenderPath(concatedPaths), dist, n, found, route, eta, isCH, nil
}

func (uc *NavigationService) ShortestPathETACH(ctx context.Context, srcLat, srcLon float64,
	dstLat float64, dstLon float64) (string, []alg.Navigation, []alg.Coordinate, float64, float64, error) {

	from := &alg.Node{
		Lat: util.TruncateFloat64(srcLat, 6),
		Lon: util.TruncateFloat64(srcLon, 6),
	}
	to := &alg.Node{
		Lat: util.TruncateFloat64(dstLat, 6),
		Lon: util.TruncateFloat64(dstLon, 6),
	}

	var err error

	fromSurakartaNode, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(from.Lat, from.Lon)
	if err != nil {
		return "", []alg.Navigation{}, []alg.Coordinate{}, 0.0, 0.0, nil
	}
	toSurakartaNode, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(to.Lat, to.Lon)
	if err != nil {
		return "", []alg.Navigation{}, []alg.Coordinate{}, 0.0, 0.0, nil
	}

	p, eta, dist := uc.CH.ShortestPathBiDijkstra(fromSurakartaNode, toSurakartaNode)

	var route []alg.Coordinate = make([]alg.Coordinate, 0)
	for n := range p {
		pathN := p[n]
		route = append(route, alg.Coordinate{
			Lat: pathN.Lat,
			Lon: pathN.Lon,
		})
	}

	n, _ := alg.CreateTurnByTurnNavigation(p)

	return alg.RenderPath(p), n, route, eta, dist, nil
}
