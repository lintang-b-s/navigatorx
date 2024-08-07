package alg

import (
	"errors"
	"fmt"
	"lintang/navigatorx/util"
	"math"
	"strings"
)

type TURN string

const (
	KIRI  TURN = "KIRI"
	KANAN TURN = "KANAN"
	LURUS TURN = "LURUS"
)

type Navigation struct {
	TurnETA     float64 `json:"turn_eta"`
	TurnDist    float64 `json:"distance_before_turn"`
	Instruction string  `json:"instruction"`
	StreetName  string  `json:"street_name"`
	Turn        TURN    `json:"turn"`
}

func CreateTurnByTurnNavigationCH(p []CHNode2) ([]Navigation, error) {
	n := []Navigation{}
	if len(p) < 4 {
		return n, nil
	}

	startSTNodeBeforeTurn := p[0]
	currStreet := p[0].StreetName
	currStDist := 0.0
	currStETA := 0.0

	for i := 0; i < len(p)-3; i++ {
		pathN2 := p[i+1]
		pathN3 := p[i+2]
		pathN4 := p[i+3]
		if currStreet != pathN3.StreetName &&
			(pathN3.StreetName != "") {

			if pathN3.StreetName != pathN4.StreetName {
				continue
			}

			// skip instruksi lewati bundaran/tugu
			if strings.Contains(pathN3.StreetName, "Bundaran") || strings.Contains(pathN3.StreetName, "Tugu") {
				continue
			}

			stNode := MakeSixDigitsAfterComa2(startSTNodeBeforeTurn, 6)
			pathN3 := MakeSixDigitsAfterComa2(pathN3, 6)
			pathN4 := MakeSixDigitsAfterComa2(pathN4, 6)

			b1 := Bearing(stNode.Lat, stNode.Lon, pathN3.Lat,
				pathN3.Lon)

			b2 := Bearing(pathN3.Lat, pathN3.Lon,
				pathN4.Lat, pathN4.Lon)

			if b1 == 0 || b2 == 0 {
				continue
			}

			turn := CalculateTurn(b1, b2)
			turnDirection := GetTurnDirection(PredictTurn(turn))

			for j := i + 4; j <= i+4+2; j++ {
				// biar turn directionnya makin akurat (ada node simpangan pathN4 yang agak gajelas)
				if j < len(p) {
					pathN5 := p[j]
					pathN5 = MakeSixDigitsAfterComa2(pathN5, 6)
					if pathN5.StreetName == pathN3.StreetName {
						b3 := Bearing(stNode.Lat, stNode.Lon, pathN3.Lat,
							pathN3.Lon)

						b4 := Bearing(pathN3.Lat, pathN3.Lon,
							pathN5.Lat, pathN5.Lon)

						if b3 == 0 || b4 == 0 {
							continue
						}

						turnAdjacentLines := CalculateTurn(b3, b4)
						turnDirectionAdjacentLines := GetTurnDirection(PredictTurn(turnAdjacentLines))

						if math.Abs(turnAdjacentLines) > math.Abs(turn) {
							turnDirection = turnDirectionAdjacentLines
						}
					}
				}

			}

			n = append(n, Navigation{
				StreetName: pathN3.StreetName,
				TurnETA:    util.RoundFloat(currStETA, 2),  //CalculateETA(startSTNodeBeforeTurn, pathN3),
				TurnDist:   util.RoundFloat(currStDist, 2), //  HaversineDistance(stLoc, turnLoc),
				Turn:       turnDirection,
			})

			startSTNodeBeforeTurn = pathN3
			currStreet = pathN3.StreetName
			currStDist = 0
			currStETA = 0
		} else {
			stLoc := NewLocation(startSTNodeBeforeTurn.Lat, startSTNodeBeforeTurn.Lon)
			pathN2Loc := NewLocation(pathN2.Lat, pathN2.Lon)
			currStDist = HaversineDistance(stLoc, pathN2Loc) * 1000
			maxSpeed := float64(30 * 1000 / 60)
			currStETA = HaversineDistance(stLoc, pathN2Loc) * 1000 / maxSpeed
		}
	}

	beforeDestionationLat := p[len(p)-1].Lat
	beforeDestionationLon := p[len(p)-1].Lon
	stLoc := NewLocation(startSTNodeBeforeTurn.Lat, startSTNodeBeforeTurn.Lon)
	pathN2Loc := NewLocation(beforeDestionationLat, beforeDestionationLon)
	currStDist = HaversineDistance(stLoc, pathN2Loc) * 1000
	maxSpeed := float64(30 * 1000 / 60)
	currStETA = HaversineDistance(stLoc, pathN2Loc) * 1000 / maxSpeed

	if len(n) == 0 {
		return []Navigation{{StreetName: "maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)"}},
			errors.New("maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)")
	}

	if n[len(n)-1].StreetName == "" {
		n[len(n)-1].StreetName = "Jalan Unknown"
	}
	n = append(n, Navigation{
		StreetName: n[len(n)-1].StreetName,
		TurnETA:    util.RoundFloat(currStETA, 2),  //CalculateETA(startSTNodeBeforeTurn, pathN3),
		TurnDist:   util.RoundFloat(currStDist, 2), //  HaversineDistance(stLoc, turnLoc),
		Turn:       LURUS,
	})

	// buat instruction
	for i := 0; i < len(n); i++ {

		if (n[i].TurnETA == 0 || n[i].TurnDist == 0) && i > 1 {
			n[i].TurnDist = n[i-1].TurnDist
			n[i].TurnETA = n[i-1].TurnETA

		}

		if i == len(n)-1 {
			n[i].Instruction = fmt.Sprintf(`LURUS dari awal %s ke tempat tujuan (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		} else if n[i].Turn != LURUS {
			n[i].Instruction = fmt.Sprintf(`Belok %s ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].Turn, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		} else {
			n[i].Instruction = fmt.Sprintf(`LURUS ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		}

		if (n[i].TurnETA == 0 || n[i].TurnDist == 0) && i > 1 {
			n = append(n[:i-1], n[i:]...)
		}
	}

	return n, nil
}

func CreateTurnByTurnNavigation(p []CHNode) ([]Navigation, error) {
	n := []Navigation{}
	if len(p) < 4 {
		return n, nil
	}

	startSTNodeBeforeTurn := p[0]
	currStreet := p[0].StreetName
	currStDist := 0.0
	currStETA := 0.0

	for i := 0; i < len(p)-3; i++ {
		pathN2 := p[i+1]
		pathN3 := p[i+2]
		pathN4 := p[i+3]
		if currStreet != pathN3.StreetName &&
			(pathN3.StreetName != "") {

			if pathN3.StreetName != pathN4.StreetName {
				continue
			}

			// skip instruksi lewati bundaran/tugu
			if strings.Contains(pathN3.StreetName, "Bundaran") || strings.Contains(pathN3.StreetName, "Tugu") {
				continue
			}

			stNode := MakeSixDigitsAfterComa(startSTNodeBeforeTurn, 6)
			pathN3 := MakeSixDigitsAfterComa(pathN3, 6)
			pathN4 := MakeSixDigitsAfterComa(pathN4, 6)

			b1 := Bearing(stNode.Lat, stNode.Lon, pathN3.Lat,
				pathN3.Lon)

			b2 := Bearing(pathN3.Lat, pathN3.Lon,
				pathN4.Lat, pathN4.Lon)

			if b1 == 0 || b2 == 0 {
				continue
			}

			turn := CalculateTurn(b1, b2)
			turnDirection := GetTurnDirection(PredictTurn(turn))

			for j := i + 4; j <= i+4+2; j++ {
				// biar turn directionnya makin akurat (ada node simpangan pathN4 yang agak gajelas)
				if j < len(p) {
					pathN5 := p[j]
					pathN5 = MakeSixDigitsAfterComa(pathN5, 6)
					if pathN5.StreetName == pathN3.StreetName {
						b3 := Bearing(stNode.Lat, stNode.Lon, pathN3.Lat,
							pathN3.Lon)

						b4 := Bearing(pathN3.Lat, pathN3.Lon,
							pathN5.Lat, pathN5.Lon)

						if b3 == 0 || b4 == 0 {
							continue
						}

						turnAdjacentLines := CalculateTurn(b3, b4)
						turnDirectionAdjacentLines := GetTurnDirection(PredictTurn(turnAdjacentLines))

						if math.Abs(turnAdjacentLines) > math.Abs(turn) {
							turnDirection = turnDirectionAdjacentLines
						}
					}
				}

			}

			n = append(n, Navigation{
				StreetName: pathN3.StreetName,
				TurnETA:    util.RoundFloat(currStETA, 2),  //CalculateETA(startSTNodeBeforeTurn, pathN3),
				TurnDist:   util.RoundFloat(currStDist, 2), //  HaversineDistance(stLoc, turnLoc),
				Turn:       turnDirection,
			})

			startSTNodeBeforeTurn = pathN3
			currStreet = pathN3.StreetName
			currStDist = 0
			currStETA = 0
		} else {
			stLoc := NewLocation(startSTNodeBeforeTurn.Lat, startSTNodeBeforeTurn.Lon)
			pathN2Loc := NewLocation(pathN2.Lat, pathN2.Lon)
			currStDist = HaversineDistance(stLoc, pathN2Loc) * 1000
			maxSpeed := float64(30 * 1000 / 60)
			currStETA = HaversineDistance(stLoc, pathN2Loc) * 1000 / maxSpeed
		}
	}

	beforeDestionationLat := p[len(p)-1].Lat
	beforeDestionationLon := p[len(p)-1].Lon
	stLoc := NewLocation(startSTNodeBeforeTurn.Lat, startSTNodeBeforeTurn.Lon)
	pathN2Loc := NewLocation(beforeDestionationLat, beforeDestionationLon)
	currStDist = HaversineDistance(stLoc, pathN2Loc) * 1000
	maxSpeed := float64(30 * 1000 / 60)
	currStETA = HaversineDistance(stLoc, pathN2Loc) * 1000 / maxSpeed

	if len(n) == 0 {
		return []Navigation{{StreetName: "maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)"}},
			errors.New("maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)")
	}

	if n[len(n)-1].StreetName == "" {
		n[len(n)-1].StreetName = "Jalan Unknown"
	}
	n = append(n, Navigation{
		StreetName: n[len(n)-1].StreetName,
		TurnETA:    util.RoundFloat(currStETA, 2),  //CalculateETA(startSTNodeBeforeTurn, pathN3),
		TurnDist:   util.RoundFloat(currStDist, 2), //  HaversineDistance(stLoc, turnLoc),
		Turn:       LURUS,
	})

	// buat instruction
	for i := 0; i < len(n); i++ {

		if (n[i].TurnETA == 0 || n[i].TurnDist == 0) && i > 1 {
			n[i].TurnDist = n[i-1].TurnDist
			n[i].TurnETA = n[i-1].TurnETA

		}

		if i == len(n)-1 {
			n[i].Instruction = fmt.Sprintf(`LURUS dari awal %s ke tempat tujuan (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		} else if n[i].Turn != LURUS {
			n[i].Instruction = fmt.Sprintf(`Belok %s ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].Turn, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		} else {
			n[i].Instruction = fmt.Sprintf(`LURUS ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
		}

		if (n[i].TurnETA == 0 || n[i].TurnDist == 0) && i > 1 {
			n = append(n[:i-1], n[i:]...)
		}
	}

	return n, nil
}

// buat semua coordinate ada 6 angka dibelakang koma
// biar itungan bearingnya ga ngaco
func MakeSixDigitsAfterComa(n CHNode, precision int) CHNode {

	if util.CountDecimalPlacesF64(n.Lat) != precision {
		n.Lat = util.RoundFloat(n.Lat+0.000001, 6)
	}
	if util.CountDecimalPlacesF64(n.Lon) != precision {
		n.Lon = util.RoundFloat(n.Lon+0.000001, 6)
	}
	return n
}

func MakeSixDigitsAfterComaLatLon(lat, lon *float64, precision int) {

	if util.CountDecimalPlacesF64(*lat) != precision {
		*lat = util.RoundFloat(*lat+0.000001, 6)
	}
	if util.CountDecimalPlacesF64(*lon) != precision {
		*lon = util.RoundFloat(*lon+0.000001, 6)
	}
}

func MakeSixDigitsAfterComa2(n CHNode2, precision int) CHNode2 {

	if util.CountDecimalPlacesF64(n.Lat) != precision {
		n.Lat = util.RoundFloat(n.Lat+0.000001, 6)
	}
	if util.CountDecimalPlacesF64(n.Lon) != precision {
		n.Lon = util.RoundFloat(n.Lon+0.000001, 6)
	}
	return n
}

func GetTurnDirection(turn string) TURN {
	switch turn {
	case "KIRI":
		return KIRI
	case "KANAN":
		return KANAN
	case "LURUS":
		return LURUS
	}

	return LURUS
}
