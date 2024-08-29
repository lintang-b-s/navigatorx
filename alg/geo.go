package alg

import (
	"math"
	"strconv"
	"strings"
)

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

// func PredictTurn(turn float64) string {
// 	if turn > 12 {
// 		return string(KANAN)
// 	} else if turn < -12 {
// 		return string(KIRI)
// 	}
// 	return string(LURUS)

// }

func calcOrientation(lat1, lon1, lat2, lon2 float64) float64 {
	shrinkFactor := math.Cos(toRadians(lat1+lat2) / 2)
	return math.Atan2(lat2-lat1, shrinkFactor*(lon2-lon1))
}

func alignOrientation(baseOrientation, orientation float64) float64 {
	var resultOrientation float64
	if baseOrientation >= 0 {
		if orientation < -math.Pi+baseOrientation {
			resultOrientation = orientation + 2*math.Pi
		} else {
			resultOrientation = orientation
		}

	} else if orientation > +math.Pi+baseOrientation {
		resultOrientation = orientation - 2*math.Pi

	} else {
		resultOrientation = orientation

	}
	return resultOrientation
}
func calculateOrientationDelta(prevLatitude, prevLongitude, latitude, longitude, prevOrientation float64) float64 {
	orientation := calcOrientation(prevLatitude, prevLongitude, latitude, longitude)
	orientation = alignOrientation(prevOrientation, orientation)
	return orientation - prevOrientation
}

func calculateSign(prevLatitude, prevLongitude, latitude, longitude, prevOrientation float64) TURN {
	delta := calculateOrientationDelta(prevLatitude, prevLongitude, latitude, longitude, prevOrientation)
	absDelta := math.Abs(delta)

	if absDelta < 0.2 {
		// 0.2 ~= 11°
		return (CONTINUE_ON_STREET)

	} else if absDelta < 0.8 {
		// 0.8 ~= 40°
		if delta > 0 {
			return (SLIGHT_LEFT)
		} else {
			return (SLIGHT_RIGHT)
		}
	} else if absDelta < 1.8 {
		// 1.8 ~= 103°
		if delta > 0 {
			return (LEFT)
		} else {
			return (RIGHT)
		}

	} else if delta > 0 {
		return (SHARP_LEFT)

	} else {
		return (SHARP_RIGHT)

	}
}

func toRadians(degrees float64) float64 {
	return degrees * 0.017453292519943295
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

func countDecimalPlaces(value float64) int {
	strValue := strconv.FormatFloat(value, 'f', -1, 64)
	parts := strings.Split(strValue, ".")

	if len(parts) < 2 {
		return 0
	}
	return len(parts[1])
}
