package alg

import "github.com/dhconnelly/rtreego"

func ProjectPointToLine(nearestStPoint Node, secondNearestStPoint Node,
	snap rtreego.Point) Coordinate {

	// proyeksi gps ke segment jalan antara 2 point tadi (ortoghonal projection)
	//a=secondNearestStPoint, b=nearestPoint, c=snap
	ab := Coordinate{nearestStPoint.Lat - secondNearestStPoint.Lat, nearestStPoint.Lon - secondNearestStPoint.Lon}
	ac := Coordinate{snap[0] - secondNearestStPoint.Lat, snap[1] - secondNearestStPoint.Lon}

	ad := (ab.Lat*ac.Lat + ab.Lon*ac.Lon) / (ab.Lat*ab.Lat + ab.Lon*ab.Lon)
	projection := Coordinate{secondNearestStPoint.Lat + ab.Lat*ad, secondNearestStPoint.Lon + ab.Lon*ad}
	return projection
}
