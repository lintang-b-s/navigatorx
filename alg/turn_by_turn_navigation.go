package alg

import (
	"errors"
	"fmt"
	"lintang/navigatorx/util"
	"strings"
)

type TURN string

const (
	LEFT               TURN = "LEFT"
	RIGHT              TURN = "RIGHT"
	CONTINUE_ON_STREET TURN = "CONTINUE_ON_STREET"
	SLIGHT_LEFT        TURN = "SLIGHT_LEFT"
	SLIGHT_RIGHT       TURN = "SLIGHT_RIGHT"
	SHARP_LEFT         TURN = "SHARP_LEFT"
	SHARP_RIGHT        TURN = "SHARP_RIGHT"
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

	currStDist := 0.0
	currStETA := 0.0
	// beforeLastN := CHNode2{}
	currStreet := p[0].StreetName
	if currStreet == "" {
		currStreet = p[1].StreetName
	}
	lastStreetNode := p[0]

	for i := 0; i < len(p)-2; i++ {
		pathN := p[i]
		pathN2 := p[i+1]
		pathN3 := p[i+2]

		startNode := NewLocation(pathN.Lat, pathN.Lon)
		twoTHNode := NewLocation(pathN2.Lat, pathN2.Lon)
		currStDist += HaversineDistance(startNode, twoTHNode) * 1000
		maxSpeed := float64(30 * 1000 / 60)
		currStETA += HaversineDistance(startNode, twoTHNode) * 1000 / maxSpeed

		// skip instruksi lewati bundaran/tugu
		if isTurnAboutCH2(pathN3) {
			continue
		}

		pathN = MakeSixDigitsAfterComa2(pathN, 6)
		pathN2 = MakeSixDigitsAfterComa2(pathN2, 6)
		pathN3 = MakeSixDigitsAfterComa2(pathN3, 6)

		b1 := calcOrientation(pathN.Lat, pathN.Lon, pathN2.Lat,
			pathN2.Lon)

		b2 := calcOrientation(pathN2.Lat, pathN2.Lon,
			pathN3.Lat, pathN3.Lon)

		if b1 == 0 || b2 == 0 {
			continue
		}

		turn := calculateSign(pathN2.Lat, pathN2.Lon,
			pathN3.Lat, pathN3.Lon, b1)

		if pathN3.StreetName == currStreet && currStreet != "" {
			lastStreetNode = pathN3
		}

		if (currStreet != "" && turn == CONTINUE_ON_STREET && pathN2.StreetName != "" && pathN3.StreetName != "" && pathN2.StreetName == pathN3.StreetName) || turn == SLIGHT_LEFT || turn == SLIGHT_RIGHT ||
			(turn == CONTINUE_ON_STREET && pathN2.StreetName == pathN3.StreetName) ||
			(currStreet != "" && currStreet == pathN3.StreetName) ||

			(turn == CONTINUE_ON_STREET && currStreet == "" && pathN2.StreetName == "" && pathN3.StreetName != "") ||
			(turn == CONTINUE_ON_STREET && pathN3.StreetName == "") ||
			(i+4 < len(p) && currStreet != pathN3.StreetName && p[i+4].StreetName != pathN3.StreetName && turn == CONTINUE_ON_STREET) ||
			(pathN2.StreetName != "" && pathN3.StreetName != "" && pathN2.StreetName == pathN3.StreetName && currStreet == pathN3.StreetName) ||
			(currStreet != "" && currStreet != pathN2.StreetName && pathN2.StreetName != pathN3.StreetName && currStreet == pathN3.StreetName) {

			continue
		}

		if currStreet != "" {
			pathN = MakeSixDigitsAfterComa2(lastStreetNode, 6)
			pathN2 = MakeSixDigitsAfterComa2(pathN2, 6)
			pathN3 = MakeSixDigitsAfterComa2(pathN3, 6)

			b1 := calcOrientation(pathN.Lat, pathN.Lon, pathN2.Lat,
				pathN2.Lon)

			b2 := calcOrientation(pathN2.Lat, pathN2.Lon,
				pathN3.Lat, pathN3.Lon)

			if b1 == 0 || b2 == 0 {
				continue
			}

			turn = calculateSign(pathN2.Lat, pathN2.Lon,
				pathN3.Lat, pathN3.Lon, b1)
		}

		if pathN3.StreetName != "" {
			currStreet = pathN3.StreetName
		}

		n = append(n, Navigation{
			StreetName: pathN3.StreetName,
			TurnETA:    util.RoundFloat(currStETA, 2),
			TurnDist:   util.RoundFloat(currStDist, 2),
			Turn:       turn,
		})

		currStDist = 0
		currStETA = 0

	}

	beforeDestionationLat := p[len(p)-1].Lat
	beforeDestionationLon := p[len(p)-1].Lon
	stLoc := NewLocation(lastStreetNode.Lat, lastStreetNode.Lon)
	pathNLoc := NewLocation(beforeDestionationLat, beforeDestionationLon)
	currStDist = HaversineDistance(stLoc, pathNLoc) * 1000
	maxSpeed := float64(30 * 1000 / 60)
	currStETA = HaversineDistance(stLoc, pathNLoc) * 1000 / maxSpeed

	if len(n) == 0 {
		return []Navigation{{StreetName: "maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)"}},
			errors.New("maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)")
	}

	// if n[len(n)-1].StreetName == "" {
	// 	n[len(n)-1].StreetName = "Jalan Unknown"
	// }
	n = append(n, Navigation{
		StreetName: n[len(n)-1].StreetName,
		TurnETA:    util.RoundFloat(currStETA, 2),  
		TurnDist:   util.RoundFloat(currStDist, 2), 
		Turn:       CONTINUE_ON_STREET,
	})

	// buat instruction
	for i := 0; i < len(n); i++ {

		if (n[i].TurnETA == 0 || n[i].TurnDist == 0) && i > 1 {
			n[i].TurnDist = n[i-1].TurnDist
			n[i].TurnETA = n[i-1].TurnETA

		}

		if i == len(n)-1 {
			if n[i].StreetName == "" {
				n[i].Instruction = fmt.Sprintf(`Lurus  ke tempat tujuan (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].TurnDist, n[i].TurnETA)
			} else {
				n[i].Instruction = fmt.Sprintf(`Lurus dari awal %s ke tempat tujuan (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
			}
		} else if n[i].Turn != CONTINUE_ON_STREET {
			if n[i].StreetName == "" {
				n[i].Instruction = fmt.Sprintf(`Belok %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].Turn, n[i].TurnDist, n[i].TurnETA)
			} else {
				n[i].Instruction = fmt.Sprintf(`Belok %s ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].Turn, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
			}
		} else {
			if n[i].StreetName == "" {
				n[i].Instruction = fmt.Sprintf(`Lurus (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].TurnDist, n[i].TurnETA)
			} else {
				n[i].Instruction = fmt.Sprintf(`Lurus ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
			}
		}

		if (n[i].TurnETA == 0 || n[i].TurnDist == 0) && i > 1 {
			n = append(n[:i-1], n[i:]...)
		}
	}

	return n, nil
}

type NodeI interface {
	CHNode | CHNode2
}

func isTurnAboutCH2(pathN CHNode2) bool {
	return strings.Contains(pathN.StreetName, "Bundaran") || strings.Contains(pathN.StreetName, "Tugu")
}

func isTurnAboutCH(pathN CHNode) bool {
	return strings.Contains(pathN.StreetName, "Bundaran") || strings.Contains(pathN.StreetName, "Tugu")
}

func CreateTurnByTurnNavigation(p []CHNode) ([]Navigation, error) {
	n := []Navigation{}
	if len(p) < 4 {
		return n, nil
	}

	// startSTNodeBeforeTurn := p[0]
	currStDist := 0.0
	currStETA := 0.0
	currStreet := p[0].StreetName
	if currStreet == "" {
		currStreet = p[1].StreetName
	}
	lastStreetNode := p[0]

	for i := 0; i < len(p)-2; i++ {
		pathN := p[i]
		pathN2 := p[i+1]
		pathN3 := p[i+2]

		startNode := NewLocation(pathN.Lat, pathN.Lon)
		twoTHNode := NewLocation(pathN2.Lat, pathN2.Lon)
		currStDist += HaversineDistance(startNode, twoTHNode) * 1000
		maxSpeed := float64(30 * 1000 / 60)
		currStETA += HaversineDistance(startNode, twoTHNode) * 1000 / maxSpeed

		// skip instruksi lewati bundaran/tugu
		if isTurnAboutCH(pathN3) {
			continue
		}

		pathN = MakeSixDigitsAfterComa(pathN, 6)
		pathN2 = MakeSixDigitsAfterComa(pathN2, 6)
		pathN3 = MakeSixDigitsAfterComa(pathN3, 6)

		b1 := calcOrientation(pathN.Lat, pathN.Lon, pathN2.Lat,
			pathN2.Lon)

		b2 := calcOrientation(pathN2.Lat, pathN2.Lon,
			pathN3.Lat, pathN3.Lon)

		if b1 == 0 || b2 == 0 {
			continue
		}

		turn := calculateSign(pathN2.Lat, pathN2.Lon,
			pathN3.Lat, pathN3.Lon, b1)

		if pathN3.StreetName == currStreet && currStreet != "" {
			lastStreetNode = pathN3
		}

		if (currStreet != "" && turn == CONTINUE_ON_STREET && pathN2.StreetName != "" && pathN3.StreetName != "" && pathN2.StreetName == pathN3.StreetName) || turn == SLIGHT_LEFT || turn == SLIGHT_RIGHT ||
			(turn == CONTINUE_ON_STREET && pathN2.StreetName == pathN3.StreetName) ||
			(currStreet != "" && currStreet == pathN3.StreetName) ||

			(turn == CONTINUE_ON_STREET && currStreet == "" && pathN2.StreetName == "" && pathN3.StreetName != "") ||
			(turn == CONTINUE_ON_STREET && pathN3.StreetName == "") ||
			(i+4 < len(p) && currStreet != pathN3.StreetName && p[i+4].StreetName != pathN3.StreetName && turn == CONTINUE_ON_STREET) ||
			(pathN2.StreetName != "" && pathN3.StreetName != "" && pathN2.StreetName == pathN3.StreetName && currStreet == pathN3.StreetName) ||
			(currStreet != "" && currStreet != pathN2.StreetName && pathN2.StreetName != pathN3.StreetName && currStreet == pathN3.StreetName) {

			continue
		}

		if currStreet != "" {
			pathN = MakeSixDigitsAfterComa(lastStreetNode, 6)
			pathN2 = MakeSixDigitsAfterComa(pathN2, 6)
			pathN3 = MakeSixDigitsAfterComa(pathN3, 6)

			b1 := calcOrientation(pathN.Lat, pathN.Lon, pathN2.Lat,
				pathN2.Lon)

			b2 := calcOrientation(pathN2.Lat, pathN2.Lon,
				pathN3.Lat, pathN3.Lon)

			if b1 == 0 || b2 == 0 {
				continue
			}

			turn = calculateSign(pathN2.Lat, pathN2.Lon,
				pathN3.Lat, pathN3.Lon, b1)
		}

		if pathN3.StreetName != "" {
			currStreet = pathN3.StreetName
		}

		n = append(n, Navigation{
			StreetName: pathN3.StreetName,
			TurnETA:    util.RoundFloat(currStETA, 2),
			TurnDist:   util.RoundFloat(currStDist, 2),
			Turn:       turn,
		})

		currStDist = 0
		currStETA = 0

	}

	beforeDestionationLat := p[len(p)-1].Lat
	beforeDestionationLon := p[len(p)-1].Lon
	stLoc := NewLocation(lastStreetNode.Lat, lastStreetNode.Lon)
	pathNLoc := NewLocation(beforeDestionationLat, beforeDestionationLon)
	currStDist = HaversineDistance(stLoc, pathNLoc) * 1000
	maxSpeed := float64(30 * 1000 / 60)
	currStETA = HaversineDistance(stLoc, pathNLoc) * 1000 / maxSpeed

	if len(n) == 0 {
		return []Navigation{{StreetName: "maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)"}},
			errors.New("maaf graph nodes dari openstreetmap  diantara shortest path route tidak ada nama jalannya (kotanya primitif)")
	}

	// if n[len(n)-1].StreetName == "" {
	// 	n[len(n)-1].StreetName = "Jalan Unknown"
	// }
	n = append(n, Navigation{
		StreetName: n[len(n)-1].StreetName,
		TurnETA:    util.RoundFloat(currStETA, 2),
		TurnDist:   util.RoundFloat(currStDist, 2),
		Turn:       CONTINUE_ON_STREET,
	})

	// buat instruction
	for i := 0; i < len(n); i++ {

		if (n[i].TurnETA == 0 || n[i].TurnDist == 0) && i > 1 {
			n[i].TurnDist = n[i-1].TurnDist
			n[i].TurnETA = n[i-1].TurnETA

		}

		if i == len(n)-1 {
			if n[i].StreetName == "" {
				n[i].Instruction = fmt.Sprintf(`Lurus  ke tempat tujuan (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].TurnDist, n[i].TurnETA)
			} else {
				n[i].Instruction = fmt.Sprintf(`Lurus dari awal %s ke tempat tujuan (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
			}
		} else if n[i].Turn != CONTINUE_ON_STREET {
			if n[i].StreetName == "" {
				n[i].Instruction = fmt.Sprintf(`Belok %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].Turn, n[i].TurnDist, n[i].TurnETA)
			} else {
				n[i].Instruction = fmt.Sprintf(`Belok %s ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].Turn, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
			}
		} else {
			if n[i].StreetName == "" {
				n[i].Instruction = fmt.Sprintf(`Lurus (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].TurnDist, n[i].TurnETA)
			} else {
				n[i].Instruction = fmt.Sprintf(`Lurus ke %s (%.2f meter dari jalan sebelumnya) (%.2f menit)`, n[i].StreetName, n[i].TurnDist, n[i].TurnETA)
			}
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
	case "LEFT":
		return LEFT
	case "RIGHT":
		return RIGHT
	case "CONTINUE_ON_STREET":
		return CONTINUE_ON_STREET
	case "SLIGHT_LEFT":
		return SLIGHT_LEFT
	case "SLIGHT_RIGHT":
		return SLIGHT_RIGHT
	case "SHARP_LEFT":
		return SHARP_LEFT
	case "SHARP_RIGHT":
		return SHARP_RIGHT
	}

	return CONTINUE_ON_STREET
}
