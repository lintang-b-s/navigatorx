package alg

import "math"

//	φ1,λ1 is the start point, φ2,λ2 the end point
//	 	φ is latitude, λ is longitude
//
// https://www.movable-type.co.uk/scripts/latlong.html
func Bearing(lat1, lon1, lat2, lon2 float64) float64 {
	p1LatRad := degToRad(lat1)
	p2LatRad := degToRad(lat2)

	diffLon := degToRad(lon2 - lon1)

	y := math.Sin(diffLon) * math.Cos(p2LatRad)
	x := math.Cos(p1LatRad)*math.Sin(p2LatRad) - math.Sin(p1LatRad)*math.Cos(p2LatRad)*math.Cos(diffLon)
	theta := math.Atan2(y, x)

	bearing := math.Mod((theta*180/math.Pi)+360, 360)
	return bearing

}

func CalculateTurn(b1, b2 float64) float64 {
	turn := b2 - b1
	if turn > 180 {
		turn -= 360
	} else if turn < -180 {
		turn += 360
	}
	return turn
}

func PredictTurn(turn float64) string {
	if turn > 13 {
		return "kanan"
	} else if turn < -13 {
		return "kiri"
	}
	return "lurus"

}

//	φ is latitude, λ is longitude
//
// https://www.movable-type.co.uk/scripts/latlong.html
func MidPoint(lat1, lon1 float64, lat2, lon2 float64) (float64, float64) {
	p1LatRad := degToRad(lat1)
	p2LatRad := degToRad(lat2)

	diffLon := degToRad(lon2 - lon1)

	bx := math.Cos(p2LatRad) * math.Cos(diffLon)
	by := math.Cos(p2LatRad) * math.Sin(diffLon)

	newLon := degToRad(lon1) + math.Atan2(by, math.Cos(p1LatRad)+bx)
	newLat := math.Atan2(math.Sin(p1LatRad)+math.Sin(p2LatRad), math.Sqrt((math.Cos(p1LatRad)+bx)*(math.Cos(p1LatRad)+bx)+by*by))

	return radToDeg(newLat), radToDeg(newLon)
}

func degToRad(d float64) float64 {
	return d * math.Pi / 180.0
}

func radToDeg(r float64) float64 {
	return 180.0 * r / math.Pi
}
