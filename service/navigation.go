package service

import (
	"context"
	"lintang/coba_osm/alg"
	"lintang/coba_osm/domain"
	"lintang/coba_osm/util"
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

	n, err := alg.CreateTurnByTurnNavigation(p)
	if err != nil {
		return alg.RenderPath(p), dist * 100, n, found, route, eta, nil
	}

	return alg.RenderPath(p), dist * 100, n, found, route, eta, nil
}

