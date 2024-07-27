package alg

import "github.com/dhconnelly/rtreego"

func ProjectPointToLine(nearestStPoint CHNode, secondNearestStPoint CHNode,
	snap rtreego.Point) Coordinate {

	// proyeksi gps ke segment jalan antara 2 point tadi (ortoghonal projection)
	//a=secondNearestStPoint, b=nearestPoint, c=snap
	ab := Coordinate{float64(nearestStPoint.Lat) - float64(secondNearestStPoint.Lat), float64(nearestStPoint.Lon) - float64(secondNearestStPoint.Lon)}
	ac := Coordinate{snap[0] - float64(secondNearestStPoint.Lat), snap[1] - float64(secondNearestStPoint.Lon)}

	ad := (ab.Lat*ac.Lat + ab.Lon*ac.Lon) / (ab.Lat*ab.Lat + ab.Lon*ab.Lon)
	projection := Coordinate{float64(secondNearestStPoint.Lat) + ab.Lat*ad, float64(secondNearestStPoint.Lon) + ab.Lon*ad}
	return projection
}
