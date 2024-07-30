package service

import (
	"context"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/domain"
	"lintang/navigatorx/util"
	"sync"
)

type ContractedGraph interface {
	SnapLocationToRoadNetworkNodeH3(ways []alg.SurakartaWay, wantToSnap []float64) int32
	ShortestPathBiDijkstra(from, to int32) ([]alg.CHNode2, float64, float64)
	AStarCH(from, to int32) (pathN []alg.CHNode, path string, eta float64, found bool, dist float64)
	IsChReady() bool
	IsCHLoaded() bool
	LoadGraph() error
	UnloadGraph() error
	IsAstarLoaded() bool
	LoadAstarGraph() error
	UnloadAstarGraph() error
}

type KVDB interface {
	GetNearestStreetsFromPointCoord(lat, lon float64) ([]alg.SurakartaWay, error)
}

type NavigationService struct {
	CH ContractedGraph
	KV KVDB
}

func NewNavigationService(contractedGraph ContractedGraph, kv KVDB) *NavigationService {
	return &NavigationService{CH: contractedGraph, KV: kv}
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
	fromSurakartaNode, err := uc.SnapLocToStreetNode(from.Lat, from.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}
	toSurakartaNode, err := uc.SnapLocToStreetNode(to.Lat, to.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}

	var pN = []alg.CHNode2{}
	var ppp = []alg.CHNode{}
	var p string
	var eta float64
	var found bool
	var dist float64
	var isCH bool
	if uc.CH.IsChReady() {

		pN, eta, dist = uc.CH.ShortestPathBiDijkstra(fromSurakartaNode, toSurakartaNode)
		p = alg.RenderPath2(pN)
		if eta != -1 {
			found = true
		}
		isCH = true
	} else {

		ppp, p, eta, found, dist = uc.CH.AStarCH(fromSurakartaNode, toSurakartaNode)
	}

	if !found {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, domain.WrapErrorf(err, domain.ErrNotFound, "sorry!! lokasi yang anda masukkan tidak tercakup di peta saya :(")
	}
	var route []alg.Coordinate = make([]alg.Coordinate, 0)

	var n = []alg.Navigation{}
	if isCH {
		n, _ = alg.CreateTurnByTurnNavigationCH(pN)
	} else {
		n, _ = alg.CreateTurnByTurnNavigation(ppp)
	}

	return p, dist, n, found, route, eta, isCH, nil
}

func (uc *NavigationService) SnapLocToStreetNode(lat, lon float64) (int32, error) {
	ways, err := uc.KV.GetNearestStreetsFromPointCoord(lat, lon)
	if err != nil {
		return 0, err
	}
	streetNodeIDx := uc.CH.SnapLocationToRoadNetworkNodeH3(ways, []float64{lat, lon})

	return streetNodeIDx, nil
}

type ShortestPathResult struct {
	PathsCH []alg.CHNode2
	Paths   []alg.CHNode
	ETA     float64
	Found   bool
	Dist    float64
	Index   int
	IsCH    bool
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
	fromSurakartaNode, err := uc.SnapLocToStreetNode(from.Lat, from.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}

	alternativeStreetSurakartaNode, err := uc.SnapLocToStreetNode(alternativeStreet.Lat, alternativeStreet.Lon)
	if err != nil {
		return "", 0, []alg.Navigation{}, false, []alg.Coordinate{}, 0.0, false, nil
	}

	toSurakartaNode, err := uc.SnapLocToStreetNode(to.Lat, to.Lon)
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
		var pN = []alg.CHNode2{}
		var ppp = []alg.CHNode{}
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
			ppp, _, eta, found, dist = uc.CH.AStarCH(fromSurakartaNode, alternativeStreetSurakartaNode)
		}
		pathChan <- ShortestPathResult{
			PathsCH: pN,
			Paths:   ppp,
			ETA:     eta,
			Found:   found,
			Dist:    dist,
			Index:   0,
			IsCH:    isCH,
		}
	}()

	go func() {

		defer wg.Done()
		var pN = []alg.CHNode2{}
		var ppp = []alg.CHNode{}
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
			ppp, _, eta, found, dist = uc.CH.AStarCH(alternativeStreetSurakartaNode, toSurakartaNode)

		}
		pathChan <- ShortestPathResult{
			PathsCH: pN,
			Paths:   ppp,
			ETA:     eta,
			Found:   found,
			Dist:    dist,
			Index:   1,
			IsCH:    isCH,
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
	concatedPathsCH := []alg.CHNode2{}
	paths[0].Paths = paths[0].Paths[:len(paths[0].Paths)-1] // exclude start node dari paths[1]
	if !paths[0].IsCH {
		concatedPaths = append(concatedPaths, paths[0].Paths...)
		concatedPaths = append(concatedPaths, paths[1].Paths...)
	} else {
		concatedPathsCH = append(concatedPathsCH, paths[0].PathsCH...)
		concatedPathsCH = append(concatedPathsCH, paths[1].PathsCH...)
	}

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

	var n = []alg.Navigation{}
	if !isCH {
		n, err = alg.CreateTurnByTurnNavigation(concatedPaths)
	} else {
		n, err = alg.CreateTurnByTurnNavigationCH(concatedPathsCH)
	}
	// var n []alg.Navigation
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

	fromSurakartaNode, err := uc.SnapLocToStreetNode(from.Lat, from.Lon)
	if err != nil {
		return "", []alg.Navigation{}, []alg.Coordinate{}, 0.0, 0.0, nil
	}
	toSurakartaNode, err := uc.SnapLocToStreetNode(to.Lat, to.Lon)
	if err != nil {
		return "", []alg.Navigation{}, []alg.Coordinate{}, 0.0, 0.0, nil
	}

	p, eta, dist := uc.CH.ShortestPathBiDijkstra(fromSurakartaNode, toSurakartaNode)

	var route []alg.Coordinate = make([]alg.Coordinate, 0)
	for n := range p {
		pathN := p[n]
		route = append(route, alg.Coordinate{
			Lat: float64(pathN.Lat),
			Lon: float64(pathN.Lon),
		})
	}

	n, _ := alg.CreateTurnByTurnNavigationCH(p)

	return alg.RenderPath2(p), n, route, eta, dist, nil
}
