package service

import (
	"context"
	"lintang/navigatorx/alg"
	"lintang/navigatorx/domain"
	"sync"
)

type ContractedGraph interface {
	SnapLocationToRoadNetworkNodeH3(ways []alg.SurakartaWay, wantToSnap []float64) int32
	ShortestPathBiDijkstra(from, to int32) ([]alg.CHNode2, float64, float64)
	AStarCH(from, to int32) (pathN []alg.CHNode, path string, eta float64, found bool, dist float64)
	SnapLocationToRoadNetworkNodeH3ForMapMatching(ways []alg.SurakartaWay, wantToSnap []float64) []alg.State
	HiddenMarkovModelMapMatching(gps []alg.StateObservationPair) []alg.CHNode2
	SnapLocationToRoadNetworkNodeRtree(lat, lon float64) (snappedRoadNodeIdx int32, err error)
	ShortestPathManyToManyBiDijkstra(from int32, to []int32) ([][]alg.CHNode2, []float64, []float64)
	ShortestPathManyToManyBiDijkstraWorkers(from []int32, to []int32) map[int32]map[int32]alg.SPSingleResultResult
	TravelingSalesmanProblemSimulatedAnnealing(cities []int32) ([]alg.CHNode2, float64, float64, [][]float64)
	IsChReady() bool
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
		Lat: srcLat,
		Lon: srcLon,
	}
	to := &alg.Node{
		Lat: dstLat,
		Lon: dstLon,
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
	// streetNodeIDx, _ := uc.CH.SnapLocationToRoadNetworkNodeRtree(lat, lon)

	return streetNodeIDx, nil
}

// func (uc *NavigationService) SnapLocToStreetNode(lat, lon float64) (int32, error) {
// 	// ways, err := uc.KV.GetNearestStreetsFromPointCoord(lat, lon)
// 	// if err != nil {
// 	// 	return 0, err
// 	// }
// 	// streetNodeIDx := uc.CH.SnapLocationToRoadNetworkNodeH3(ways, []float64{lat, lon})
// 	streetNodeIDx, _ := uc.CH.SnapLocationToRoadNetworkNodeRtree(lat, lon)

// 	return streetNodeIDx, nil
// }

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
		Lat: srcLat,
		Lon: srcLon,
	}

	alternativeStreet := &alg.Node{
		Lat: alternativeStreetLat,
		Lon: alternativeStreetLon,
	}

	to := &alg.Node{
		Lat: dstLat,
		Lon: dstLon,
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

	for i := 0; i < 2; i++ {

	}

	go func(wgg *sync.WaitGroup) {

		defer wgg.Done()
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
	}(&wg)

	go func(wgg *sync.WaitGroup) {

		defer wgg.Done()
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

	}(&wg)

	go func() {
		wg.Wait()
		close(pathChan)
	}()

	for p := range pathChan {
		if p.Index == 0 {
			paths[0] = p
		} else {
			paths[1] = p
		}
	}

	concatedPaths := []alg.CHNode{}
	concatedPathsCH := []alg.CHNode2{}
	if !paths[0].IsCH {
		paths[0].Paths = paths[0].Paths[:len(paths[0].Paths)-1] // exclude start node dari paths[1]
		concatedPaths = append(concatedPaths, paths[0].Paths...)
		concatedPaths = append(concatedPaths, paths[1].Paths...)
	} else {
		paths[0].PathsCH = paths[0].PathsCH[:len(paths[0].PathsCH)-1] // exclude start node dari paths[1]
		concatedPathsCH = append(concatedPathsCH, paths[0].PathsCH...)
		concatedPathsCH = append(concatedPathsCH, paths[1].PathsCH...)
	}

	eta := paths[0].ETA + paths[1].ETA
	dist := paths[0].Dist + paths[1].Dist
	found := paths[0].Found && paths[1].Found
	isCH := paths[0].IsCH
	// eta satuannya minute
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
	if err != nil {
		return alg.RenderPath(concatedPaths), dist, n, found, route, eta, isCH, nil
	}

	return alg.RenderPath(concatedPaths), dist, n, found, route, eta, isCH, nil
}

func (uc *NavigationService) ShortestPathETACH(ctx context.Context, srcLat, srcLon float64,
	dstLat float64, dstLon float64) (string, []alg.Navigation, []alg.Coordinate, float64, float64, error) {

	from := &alg.Node{
		Lat: srcLat,
		Lon: srcLon,
	}
	to := &alg.Node{
		Lat: dstLat,
		Lon: dstLon,
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
			Lat: pathN.Lat,
			Lon: pathN.Lon,
		})
	}

	n, _ := alg.CreateTurnByTurnNavigationCH(p)

	return alg.RenderPath2(p), n, route, eta, dist, nil
}

func (uc *NavigationService) HiddenMarkovModelMapMatching(ctx context.Context, gps []alg.Coordinate) (string, []alg.CHNode2, error) {
	hmmPair := []alg.StateObservationPair{}

	stateID := 0
	for i, gpsPoint := range gps {
		if i < len(gps)-1 && len(gps) > 300 {
			// preprocessing , buang gps points yang jaraknya lebih dari 2*4.07 meter dari previous gps point
			// (Hidden Markov Map Matching Through Noise and Sparseness 4.1) , biar gak terlalu lama viterbinya O(T*|S|^2)
			currGpsLoc := alg.NewLocation(gpsPoint.Lat, gpsPoint.Lon)
			nextGpsLoc := alg.NewLocation(gps[i+1].Lat, gps[i+1].Lon)
			if alg.HaversineDistance(currGpsLoc, nextGpsLoc)*1000 >= 2*4.07 {
				continue
			}
		}
		nearestRoadNodes, err := uc.NearestStreetNodesForMapMatching(gpsPoint.Lat, gpsPoint.Lon)
		if len(nearestRoadNodes) == 0 {
			continue
		}
		if err != nil {
			return "", []alg.CHNode2{}, err
		}
		for i := range nearestRoadNodes {
			nearestRoadNodes[i].ID = stateID

			stateID++
		}

		chNodeGPS := alg.CHNode2{
			Lat: gpsPoint.Lat,
			Lon: gpsPoint.Lon,
		}

		hmmPair = append(hmmPair, alg.StateObservationPair{
			Observation: chNodeGPS,
			State:       nearestRoadNodes,
		})
	}

	path := uc.CH.HiddenMarkovModelMapMatching(hmmPair)
	return alg.RenderPath2(path), path, nil
}

type TargetResult struct {
	TargetCoord alg.Coordinate
	Path        string
	Dist        float64
	ETA         float64
}

func (uc *NavigationService) ManyToManyQuery(ctx context.Context, sourcesLat, sourcesLon, destsLat, destsLon []float64) map[alg.Coordinate][]TargetResult {
	sources := []int32{}
	dests := []int32{}

	for i := 0; i < len(sourcesLat); i++ {
		srcNode, _ := uc.SnapLocToStreetNode(sourcesLat[i], sourcesLon[i])
		sources = append(sources, srcNode)
	}

	for i := 0; i < len(destsLat); i++ {
		dstNode, _ := uc.SnapLocToStreetNode(destsLat[i], destsLon[i])
		dests = append(dests, dstNode)
	}

	manyToManyRes := make(map[alg.Coordinate][]TargetResult)

	scMap := uc.CH.ShortestPathManyToManyBiDijkstraWorkers(sources, dests)

	for i, src := range sources {
		srcCoord := alg.Coordinate{
			Lat: sourcesLat[i],
			Lon: sourcesLon[i],
		}

		for j, dest := range dests {
			currPath := alg.RenderPath2(scMap[src][dest].Paths)
			currDist := scMap[src][dest].Dist
			currETA := scMap[src][dest].Eta

			manyToManyRes[srcCoord] = append(manyToManyRes[srcCoord], TargetResult{
				TargetCoord: alg.Coordinate{
					Lat: destsLat[j],
					Lon: destsLon[j],
				},
				Path: currPath,
				Dist: currDist,
				ETA:  currETA,
			})
		}

	}
	return manyToManyRes
}

func (uc *NavigationService) TravelingSalesmanProblemSimulatedAnneal(ctx context.Context, citiesLat []float64, citiesLon []float64) ([]alg.Coordinate, string, float64, float64) {

	citiesID := []int32{}
	for i := 0; i < len(citiesLat); i++ {
		cityNode, _ := uc.SnapLocToStreetNode(citiesLat[i], citiesLon[i])
		citiesID = append(citiesID, cityNode)
	}

	tspTourNodes, bestETA, bestDistance, bestTourCitiesOrder := uc.CH.TravelingSalesmanProblemSimulatedAnnealing(citiesID)
	cititesTour := []alg.Coordinate{}
	for i := 0; i < len(bestTourCitiesOrder); i++ {
		cititesTour = append(cititesTour, alg.Coordinate{
			Lat: bestTourCitiesOrder[i][0],
			Lon: bestTourCitiesOrder[i][1],
		})
	}
	return cititesTour, alg.RenderPath2(tspTourNodes), bestETA, bestDistance
}

func (uc *NavigationService) NearestStreetNodesForMapMatching(lat, lon float64) ([]alg.State, error) {
	ways, err := uc.KV.GetNearestStreetsFromPointCoord(lat, lon)
	if err != nil {
		return []alg.State{}, err
	}
	streetNodes := uc.CH.SnapLocationToRoadNetworkNodeH3ForMapMatching(ways, []float64{lat, lon})
	// streetNodes, err := uc.CH.SnapLocationToRoadNetworkNodeRtree(lat, lon) // gakjadi pake rtree
	if err != nil {
		return []alg.State{}, err
	}
	return streetNodes, nil
}
