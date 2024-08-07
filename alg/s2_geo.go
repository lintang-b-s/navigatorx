package alg

import (
	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s2"
)

func ProjectPointToLineCoord(nearestStPoint CHNode2, secondNearestStPoint CHNode2,
	snap rtreego.Point) Coordinate {
	nearestStPoint = MakeSixDigitsAfterComa2(nearestStPoint, 6)
	secondNearestStPoint = MakeSixDigitsAfterComa2(secondNearestStPoint, 6)
	snapLat := snap[0]
	snapLon := snap[1]
	MakeSixDigitsAfterComaLatLon(&snapLat, &snapLon, 6)

	nearestStS2 := s2.PointFromLatLng(s2.LatLngFromDegrees(nearestStPoint.Lat, nearestStPoint.Lon))
	secondNearestStS2 := s2.PointFromLatLng(s2.LatLngFromDegrees(secondNearestStPoint.Lat, secondNearestStPoint.Lon))
	snapS2 := s2.PointFromLatLng(s2.LatLngFromDegrees(snapLat, snapLon))
	projection := s2.Project(snapS2, nearestStS2, secondNearestStS2)
	projectLatLng := s2.LatLngFromPoint(projection)
	return Coordinate{projectLatLng.Lat.Degrees(), projectLatLng.Lng.Degrees()}
}
